package clients_test

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	uuid "github.com/hashicorp/go-uuid"

	"lockbox.dev/clients"
	"lockbox.dev/clients/storers/memory"
	"lockbox.dev/clients/storers/postgres"
)

const (
	changeSecret = 1 << iota
	changeName
	changeVariations
)

var factories []Factory

type Factory interface {
	NewStorer(ctx context.Context) (clients.Storer, error)
	TeardownStorers() error
}

func uuidOrFail(t *testing.T) string {
	t.Helper()
	id, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatalf("Unexpected error generating ID: %s", err.Error())
	}
	return id
}

func TestMain(m *testing.M) {
	flag.Parse()

	// set up our test storers
	factories = append(factories, memory.Factory{})
	if os.Getenv(postgres.TestConnStringEnvVar) != "" {
		storerConn, err := sql.Open("postgres", os.Getenv(postgres.TestConnStringEnvVar))
		if err != nil {
			panic(err)
		}
		factories = append(factories, postgres.NewFactory(storerConn))
	}

	// run the tests
	result := m.Run()

	// tear down all the storers we created
	for _, factory := range factories {
		err := factory.TeardownStorers()
		if err != nil {
			log.Printf("Error cleaning up after %T: %s", factory, err.Error())
		}
	}

	// return the test result
	os.Exit(result)
}

func runTest(t *testing.T, testFunc func(*testing.T, clients.Storer, context.Context)) {
	t.Helper()
	for _, factory := range factories {
		ctx := context.Background()
		storer, err := factory.NewStorer(ctx)
		if err != nil {
			t.Fatalf("Error creating Storer from %T: %s", factory, err.Error())
		}
		t.Run(fmt.Sprintf("Storer=%T", storer), func(t *testing.T) {
			t.Parallel()
			testFunc(t, storer, ctx)
		})
	}
}

func TestClientCreateGetDelete(t *testing.T) {
	t.Parallel()

	runTest(t, func(t *testing.T, storer clients.Storer, ctx context.Context) {
		client := clients.Client{
			ID:           uuidOrFail(t),
			Name:         "Test Client",
			Confidential: true,
			CreatedAt:    time.Now().Round(time.Millisecond),
			CreatedBy:    "test",
			CreatedByIP:  "127.0.0.1",
		}
		ch, err := clients.ChangeSecret([]byte("test secret"))
		if err != nil {
			t.Fatalf("Error generating client secret: %s", err)
		}
		client = clients.Apply(ch, client)
		err = storer.Create(ctx, client)
		if err != nil {
			t.Fatalf("Error creating client: %s", err)
		}
		res, err := storer.Get(ctx, client.ID)
		if err != nil {
			t.Errorf("Error retrieving client: %s", err)
		}
		if diff := cmp.Diff(client, res); diff != "" {
			t.Errorf("Unexpected diff (-wanted, +got): %s", diff)
		}
		err = storer.Delete(ctx, client.ID)
		if err != nil {
			t.Errorf("Error deleting client: %s", err)
		}
		_, err = storer.Get(ctx, client.ID)
		if !errors.Is(err, clients.ErrClientNotFound) {
			t.Errorf("Expected %v, got %v instead", clients.ErrClientNotFound, err)
		}
	})
}

func TestClientUpdateOneOfMany(t *testing.T) {
	t.Parallel()

	runTest(t, func(t *testing.T, storer clients.Storer, ctx context.Context) {
		for variation := 1; variation < changeVariations; variation++ {
			variation := variation
			t.Run(fmt.Sprintf("variation=%d", variation), func(t *testing.T) {
				t.Parallel()

				client := clients.Client{
					ID:           uuidOrFail(t),
					Name:         "Test Client",
					Confidential: true,
					CreatedAt:    time.Now().Round(time.Millisecond),
					CreatedBy:    "test",
					CreatedByIP:  "127.0.0.1",
				}
				change, err := clients.ChangeSecret([]byte("test secret"))
				if err != nil {
					t.Fatalf("Error generating client secret: %s", err)
				}
				client = clients.Apply(change, client)
				err = storer.Create(ctx, client)
				if err != nil {
					t.Errorf("Error creating client: %s", err)
				}

				var throwaways []clients.Client
				for idx := 0; idx < 5; idx++ {
					throwaways = append(throwaways, clients.Client{
						ID:           uuidOrFail(t),
						Name:         fmt.Sprintf("Test Client %d", variation),
						Confidential: variation%2 == 0,
						CreatedAt:    time.Now().Round(time.Millisecond),
						CreatedBy:    "test",
						CreatedByIP:  "127.0.0.1",
					})
					change, err = clients.ChangeSecret([]byte("test secret " + client.ID))
					if err != nil {
						t.Fatalf("Error generating client secret: %s", err)
					}
					throwaways[idx] = clients.Apply(change, throwaways[idx])
					err = storer.Create(ctx, throwaways[idx])
					if err != nil {
						t.Errorf("Error creating throwaway: %v", err)
					}
				}
				change = clients.Change{}
				if variation&changeSecret != 0 {
					var secretChange clients.Change
					secretChange, err = clients.ChangeSecret([]byte("changed secret"))
					if err != nil {
						t.Errorf("Error generating client secret: %s", err)
					}
					change.SecretHash = secretChange.SecretHash
					change.SecretScheme = secretChange.SecretScheme
				}
				if variation&changeName != 0 {
					name := fmt.Sprintf("Updated Test Client %d", variation)
					change.Name = &name
				}
				expectation := clients.Apply(change, client)
				err = storer.Update(ctx, client.ID, change)
				if err != nil {
					t.Errorf("Unexpected error updating client: %v", err)
				}
				result, err := storer.Get(ctx, client.ID)
				if err != nil {
					t.Errorf("Unexpected error retrieving client: %v", err)
				}
				if diff := cmp.Diff(expectation, result); diff != "" {
					t.Errorf("Unexpected diff (-wanted, +got): %s", diff)
				}
				for _, throwaway := range throwaways {
					result, err := storer.Get(ctx, throwaway.ID)
					if err != nil {
						t.Errorf("Unexpected error retrieving client: %v", err)
					}
					if diff := cmp.Diff(throwaway, result); diff != "" {
						t.Errorf("Unexpected diff (-wanted, +got): %s", diff)
					}
				}
			})
		}
	})
}

func TestClientUpdateNoChange(t *testing.T) {
	t.Parallel()

	// updating an account with an empty change should not error
	runTest(t, func(t *testing.T, storer clients.Storer, ctx context.Context) {
		client := clients.Client{
			ID:           uuidOrFail(t),
			Name:         "Test Client",
			Confidential: true,
			CreatedAt:    time.Now().Round(time.Millisecond),
			CreatedBy:    "test",
			CreatedByIP:  "127.0.0.1",
		}
		ch, err := clients.ChangeSecret([]byte("test secret"))
		if err != nil {
			t.Fatalf("Error generating client secret: %s", err)
		}
		client = clients.Apply(ch, client)
		err = storer.Create(ctx, client)
		if err != nil {
			t.Errorf("Error creating client: %s", err)
		}
		var change clients.Change
		err = storer.Update(ctx, client.ID, change)
		if err != nil {
			t.Fatalf("Unexpected error updating client: %+v\n", err)
		}
	})
}

func TestClientAlreadyExists(t *testing.T) {
	t.Parallel()

	runTest(t, func(t *testing.T, storer clients.Storer, ctx context.Context) {
		client := clients.Client{
			ID:           uuidOrFail(t),
			Name:         "Test Client",
			Confidential: true,
			CreatedAt:    time.Now().Round(time.Millisecond),
			CreatedBy:    "test",
			CreatedByIP:  "127.0.0.1",
		}
		ch, err := clients.ChangeSecret([]byte("test secret"))
		if err != nil {
			t.Fatalf("Error generating client secret: %s", err)
		}
		client = clients.Apply(ch, client)
		err = storer.Create(ctx, client)
		if err != nil {
			t.Fatalf("Error creating client: %s", err)
		}
		err = storer.Create(ctx, client)
		if !errors.Is(err, clients.ErrClientAlreadyExists) {
			t.Errorf("Expected %v, got %v", clients.ErrClientAlreadyExists, err)
		}
	})
}

func TestClientGetNonexistent(t *testing.T) {
	t.Parallel()

	runTest(t, func(t *testing.T, storer clients.Storer, ctx context.Context) {
		_, err := storer.Get(ctx, "nope")
		if !errors.Is(err, clients.ErrClientNotFound) {
			t.Fatalf("Expected %v, got %v instead", clients.ErrClientNotFound, err)
		}
	})
}

func TestClientUpdateNonexistent(t *testing.T) {
	t.Parallel()

	runTest(t, func(t *testing.T, storer clients.Storer, ctx context.Context) {
		ch, err := clients.ChangeSecret([]byte("test secret"))
		if err != nil {
			t.Fatalf("Error generating client secret: %s", err)
		}
		err = storer.Update(ctx, uuidOrFail(t), ch)
		if err != nil {
			t.Fatalf("Expected %v, got %v instead", nil, err)
		}
	})
}

func TestClientDeleteNonexistent(t *testing.T) {
	t.Parallel()

	runTest(t, func(t *testing.T, storer clients.Storer, ctx context.Context) {
		err := storer.Delete(ctx, uuidOrFail(t))
		if err != nil {
			t.Fatalf("Expected %v, got %v instead", nil, err)
		}
	})
}

func TestRedirectURIsCreateListDelete(t *testing.T) {
	t.Parallel()

	runTest(t, func(t *testing.T, storer clients.Storer, ctx context.Context) {
		client := clients.Client{
			ID:           uuidOrFail(t),
			Name:         "Test Client",
			Confidential: true,
			CreatedAt:    time.Now().Round(time.Millisecond),
			CreatedBy:    "test",
			CreatedByIP:  "127.0.0.1",
		}
		ch, err := clients.ChangeSecret([]byte("test secret"))
		if err != nil {
			t.Fatalf("Error generating client secret: %s", err)
		}
		client = clients.Apply(ch, client)
		err = storer.Create(ctx, client)
		if err != nil {
			t.Fatalf("Error creating client: %s", err)
		}

		redirectURIs := []clients.RedirectURI{}
		// add URIs in 4 separate groups, with 1, 2, 3, and 4 URIs in each group
		// this checks that listing URIs when they're added over time works
		for group := 1; group < 5; group++ {
			newRedirectURIs := []clients.RedirectURI{}
			for uri := 0; uri < group; uri++ {
				newRedirectURIs = append(newRedirectURIs, clients.RedirectURI{
					ID:          uuidOrFail(t),
					URI:         fmt.Sprintf("https://test-%d-%d.impractical.services/testing", group, uri),
					IsBaseURI:   (group+uri)%2 == 0,
					ClientID:    client.ID,
					CreatedAt:   time.Now().Round(time.Millisecond),
					CreatedBy:   "test",
					CreatedByIP: "127.0.0.1",
				})
			}
			err = storer.AddRedirectURIs(ctx, newRedirectURIs)
			if err != nil {
				t.Errorf("Error storing redirect URIs: %s", err)
			}
			redirectURIs = append(redirectURIs, newRedirectURIs...)

			var res []clients.RedirectURI
			res, err = storer.ListRedirectURIs(ctx, client.ID)
			if err != nil {
				t.Errorf("Error retrieving redirect URIs: %s", err)
			}
			clients.RedirectURIsByURI(redirectURIs)
			if len(redirectURIs) != len(res) {
				t.Fatalf("Expected %d results, got %d", len(redirectURIs), len(res))
			}
			for pos, uri := range redirectURIs {
				if diff := cmp.Diff(uri, res[pos]); diff != "" {
					t.Errorf("Unexpected diff (-wanted, +got): %s", diff)
				}
			}
		}
		err = storer.RemoveRedirectURIs(ctx, []string{redirectURIs[0].ID})
		if err != nil {
			t.Errorf("Error removing redirect URIs: %s", err)
		}
		res, err := storer.ListRedirectURIs(ctx, client.ID)
		if err != nil {
			t.Errorf("Error retrieving redirect URIs: %s", err)
		}
		clients.RedirectURIsByURI(res)
		if len(redirectURIs[1:]) != len(res) {
			t.Errorf("Expected %d results, got %d", len(redirectURIs[1:]), len(res))
		}
		ids := make([]string, 0, len(redirectURIs[1:]))
		for pos, uri := range redirectURIs[1:] {
			if diff := cmp.Diff(uri, res[pos]); diff != "" {
				t.Errorf("Unexpected diff (-wanted, +got): %s", diff)
			}
			ids = append(ids, uri.ID)
		}
		err = storer.RemoveRedirectURIs(ctx, ids)
		if err != nil {
			t.Errorf("Error removing redirect URIs: %v", err)
		}
		res, err = storer.ListRedirectURIs(ctx, client.ID)
		if err != nil {
			t.Errorf("Error retrieving redirect URIs: %s", err)
		}
		if len(res) != 0 {
			t.Errorf("Expected no results, got %v", res)
		}
	})
}

func TestRedirectURIsListNone(t *testing.T) {
	t.Parallel()

	runTest(t, func(t *testing.T, storer clients.Storer, ctx context.Context) {
		client := clients.Client{
			ID:           uuidOrFail(t),
			Name:         "Test Client",
			Confidential: true,
			CreatedAt:    time.Now().Round(time.Millisecond),
			CreatedBy:    "test",
			CreatedByIP:  "127.0.0.1",
		}
		ch, err := clients.ChangeSecret([]byte("test secret"))
		if err != nil {
			t.Fatalf("Error generating client secret: %s", err)
		}
		client = clients.Apply(ch, client)
		err = storer.Create(ctx, client)
		if err != nil {
			t.Fatalf("Error creating client: %s", err)
		}
		res, err := storer.ListRedirectURIs(ctx, client.ID)
		if err != nil {
			t.Errorf("Error retrieving redirect URIs: %s", err)
		}
		if len(res) != 0 {
			t.Errorf("Expected no results, got %v", res)
		}
	})
}

func TestRedirectURIsListNonexistantClient(t *testing.T) {
	t.Parallel()

	runTest(t, func(t *testing.T, storer clients.Storer, ctx context.Context) {
		res, err := storer.ListRedirectURIs(ctx, uuidOrFail(t))
		if err != nil {
			t.Errorf("Error retrieving redirect URIs: %s", err)
		}
		if len(res) != 0 {
			t.Errorf("Expected no results, got %v", res)
		}
	})
}

func TestRedirectURIIDAlreadyExists(t *testing.T) {
	t.Parallel()

	runTest(t, func(t *testing.T, storer clients.Storer, ctx context.Context) {
		client := clients.Client{
			ID:           uuidOrFail(t),
			Name:         "Test Client",
			Confidential: true,
			CreatedAt:    time.Now().Round(time.Millisecond),
			CreatedBy:    "test",
			CreatedByIP:  "127.0.0.1",
		}
		ch, err := clients.ChangeSecret([]byte("test secret"))
		if err != nil {
			t.Fatalf("Error generating client secret: %s", err)
		}
		client = clients.Apply(ch, client)
		err = storer.Create(ctx, client)
		if err != nil {
			t.Fatalf("Error creating client: %s", err)
		}

		uri := clients.RedirectURI{
			ID:          uuidOrFail(t),
			URI:         "https://test-1.impractical.services/testing",
			IsBaseURI:   false,
			ClientID:    client.ID,
			CreatedAt:   time.Now().Round(time.Millisecond),
			CreatedBy:   "test",
			CreatedByIP: "127.0.0.1",
		}
		err = storer.AddRedirectURIs(ctx, []clients.RedirectURI{uri})
		if err != nil {
			t.Fatalf("Error adding redirect URI: %s", err)
		}
		uri.URI += "/test"
		uri2 := clients.RedirectURI{
			ID:          uuidOrFail(t),
			URI:         "https://test-2.impractical.services/testing",
			IsBaseURI:   false,
			ClientID:    client.ID,
			CreatedAt:   time.Now().Round(time.Millisecond),
			CreatedBy:   "test",
			CreatedByIP: "127.0.0.1",
		}
		err = storer.AddRedirectURIs(ctx, []clients.RedirectURI{uri, uri2})
		var redirectURIError clients.RedirectURIAlreadyExistsError
		if ok := errors.As(err, &redirectURIError); !ok {
			t.Errorf("Expected %T, got %v", clients.RedirectURIAlreadyExistsError{}, err)
		} else if uri.ID != redirectURIError.ID {
			t.Errorf("Expected RedirectURIAlreadyExistsError to be for %s, was for %s", uri.ID, redirectURIError.ID)
		}
	})
}

func TestRedirectURIURIAlreadyExists(t *testing.T) {
	t.Parallel()

	runTest(t, func(t *testing.T, storer clients.Storer, ctx context.Context) {
		client := clients.Client{
			ID:           uuidOrFail(t),
			Name:         "Test Client",
			Confidential: true,
			CreatedAt:    time.Now().Round(time.Millisecond),
			CreatedBy:    "test",
			CreatedByIP:  "127.0.0.1",
		}
		ch, err := clients.ChangeSecret([]byte("test secret"))
		if err != nil {
			t.Fatalf("Error generating client secret: %s", err)
		}
		client = clients.Apply(ch, client)
		err = storer.Create(ctx, client)
		if err != nil {
			t.Fatalf("Error creating client: %s", err)
		}

		uri := clients.RedirectURI{
			ID:          uuidOrFail(t),
			URI:         "https://test-1.impractical.services/testing",
			IsBaseURI:   false,
			ClientID:    client.ID,
			CreatedAt:   time.Now().Round(time.Millisecond),
			CreatedBy:   "test",
			CreatedByIP: "127.0.0.1",
		}
		err = storer.AddRedirectURIs(ctx, []clients.RedirectURI{uri})
		if err != nil {
			t.Fatalf("Error adding redirect URI: %s", err)
		}
		uri.ID = uuidOrFail(t)
		uri2 := clients.RedirectURI{
			ID:          uuidOrFail(t),
			URI:         "https://test-2.impractical.services/testing",
			IsBaseURI:   false,
			ClientID:    client.ID,
			CreatedAt:   time.Now().Round(time.Millisecond),
			CreatedBy:   "test",
			CreatedByIP: "127.0.0.1",
		}
		err = storer.AddRedirectURIs(ctx, []clients.RedirectURI{uri, uri2})
		var redirectURIError clients.RedirectURIAlreadyExistsError
		if ok := errors.As(err, &redirectURIError); !ok {
			t.Errorf("Expected %T, got %v", clients.RedirectURIAlreadyExistsError{}, err)
		} else if uri.URI != redirectURIError.URI {
			t.Errorf("Expected RedirectURIAlreadyExistsError to be for %s, was for %s", uri.URI, redirectURIError.URI)
		}
	})
}

func TestRedirectURIDeleteNonexistent(t *testing.T) {
	t.Parallel()

	runTest(t, func(t *testing.T, storer clients.Storer, ctx context.Context) {
		err := storer.RemoveRedirectURIs(ctx, []string{uuidOrFail(t)})
		if err != nil {
			t.Fatalf("Expected %v, got %v instead", nil, err)
		}
	})
}
