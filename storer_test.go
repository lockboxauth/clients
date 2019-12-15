package clients_test

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

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

func compareClients(client1, client2 clients.Client) (ok bool, field string, val1, val2 interface{}) {
	if client1.ID != client2.ID {
		return false, "ID", client1.ID, client2.ID
	}
	if client1.Name != client2.Name {
		return false, "Name", client1.Name, client2.Name
	}
	if client1.SecretHash != client2.SecretHash {
		return false, "SecretHash", client1.SecretHash, client2.SecretHash
	}
	if client1.SecretScheme != client2.SecretScheme {
		return false, "Scheme", client1.SecretScheme, client2.SecretScheme
	}
	if client1.Confidential != client2.Confidential {
		return false, "Confidential", client1.Confidential, client2.Confidential
	}
	if !client1.CreatedAt.Equal(client2.CreatedAt) {
		return false, "CreatedAt", client1.CreatedAt, client2.CreatedAt
	}
	if client1.CreatedBy != client2.CreatedBy {
		return false, "CreatedBy", client1.CreatedBy, client2.CreatedBy
	}
	if client1.CreatedByIP != client2.CreatedByIP {
		return false, "CreatedByIP", client1.CreatedByIP, client2.CreatedByIP
	}
	return true, "", nil, nil
}

func compareRedirectURIs(uri1, uri2 clients.RedirectURI) (ok bool, field string, val1, val2 interface{}) {
	if uri1.ID != uri2.ID {
		return false, "ID", uri1.ID, uri2.ID
	}
	if uri1.URI != uri2.URI {
		return false, "URI", uri1.URI, uri2.URI
	}
	if uri1.IsBaseURI != uri2.IsBaseURI {
		return false, "IsBaseURI", uri1.IsBaseURI, uri2.IsBaseURI
	}
	if uri1.ClientID != uri2.ClientID {
		return false, "ClientID", uri1.ClientID, uri2.ClientID
	}
	if !uri1.CreatedAt.Equal(uri2.CreatedAt) {
		return false, "CreatedAt", uri1.CreatedAt, uri2.CreatedAt
	}
	if uri1.CreatedBy != uri2.CreatedBy {
		return false, "CreatedBy", uri1.CreatedBy, uri2.CreatedBy
	}
	if uri1.CreatedByIP != uri2.CreatedByIP {
		return false, "CreatedByIP", uri1.CreatedByIP, uri2.CreatedByIP
	}
	return true, "", nil, nil
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

func runTest(t *testing.T, f func(*testing.T, clients.Storer, context.Context)) {
	t.Parallel()
	for _, factory := range factories {
		ctx := context.Background()
		storer, err := factory.NewStorer(ctx)
		if err != nil {
			t.Fatalf("Error creating Storer from %T: %s", factory, err.Error())
		}
		t.Run(fmt.Sprintf("Storer=%T", storer), func(t *testing.T) {
			t.Parallel()
			f(t, storer, ctx)
		})
	}
}

func TestClientCreateGetDelete(t *testing.T) {
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
		ok, field, expected, got := compareClients(client, res)
		if !ok {
			t.Errorf("Expected %v for %q, got %v", expected, field, got)
		}
		err = storer.Delete(ctx, client.ID)
		if err != nil {
			t.Errorf("Error deleting client: %s", err)
		}
		_, err = storer.Get(ctx, client.ID)
		if err != clients.ErrClientNotFound {
			t.Errorf("Expected %v, got %v instead", clients.ErrClientNotFound, err)
		}
	})
}

func TestClientUpdateOneOfMany(t *testing.T) {
	runTest(t, func(t *testing.T, storer clients.Storer, ctx context.Context) {
		for i := 1; i < changeVariations; i++ {
			i := i
			t.Run(fmt.Sprintf("i=%d", i), func(t *testing.T) {
				t.Parallel()

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

				var throwaways []clients.Client
				for x := 0; x < 5; x++ {
					throwaways = append(throwaways, clients.Client{
						ID:           uuidOrFail(t),
						Name:         fmt.Sprintf("Test Client %d", i),
						Confidential: i%2 == 0,
						CreatedAt:    time.Now().Round(time.Millisecond),
						CreatedBy:    "test",
						CreatedByIP:  "127.0.0.1",
					})
					ch, err := clients.ChangeSecret([]byte("test secret " + client.ID))
					if err != nil {
						t.Fatalf("Error generating client secret: %s", err)
					}
					throwaways[x] = clients.Apply(ch, throwaways[x])
					err = storer.Create(ctx, throwaways[x])
					if err != nil {
						t.Errorf("Error creating throwaway: %v", err)
					}
				}
				var change clients.Change
				if i&changeSecret != 0 {
					ch, err = clients.ChangeSecret([]byte("changed secret"))
					if err != nil {
						t.Errorf("Error generating client secret: %s", err)
					}
					change.SecretHash = ch.SecretHash
					change.SecretScheme = ch.SecretScheme
				}
				if i&changeName != 0 {
					name := fmt.Sprintf("Updated Test Client %d", i)
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
				ok, field, exp, res := compareClients(expectation, result)
				if !ok {
					t.Errorf("Expected %s to be %v, got %v", field, exp, res)
				}
				for _, throwaway := range throwaways {
					result, err := storer.Get(ctx, throwaway.ID)
					if err != nil {
						t.Errorf("Unexpected error retrieving client: %v", err)
					}
					ok, field, exp, res := compareClients(throwaway, result)
					if !ok {
						t.Errorf("Expected %s to be %v, got %v", field, exp, res)
					}
				}
			})
		}
	})
}

func TestClientUpdateNoChange(t *testing.T) {
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
		if err != clients.ErrClientAlreadyExists {
			t.Errorf("Expected %v, got %v", clients.ErrClientAlreadyExists, err)
		}
	})
}

func TestClientGetNonexistent(t *testing.T) {
	runTest(t, func(t *testing.T, storer clients.Storer, ctx context.Context) {
		_, err := storer.Get(ctx, "nope")
		if err != clients.ErrClientNotFound {
			t.Fatalf("Expected %v, got %v instead", clients.ErrClientNotFound, err)
		}
	})
}

func TestClientUpdateNonexistent(t *testing.T) {
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
	runTest(t, func(t *testing.T, storer clients.Storer, ctx context.Context) {
		err := storer.Delete(ctx, uuidOrFail(t))
		if err != nil {
			t.Fatalf("Expected %v, got %v instead", nil, err)
		}
	})
}

func TestRedirectURIsCreateListDelete(t *testing.T) {
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
		for x := 1; x < 5; x++ {
			newRedirectURIs := []clients.RedirectURI{}
			for y := 0; y < x; y++ {
				newRedirectURIs = append(newRedirectURIs, clients.RedirectURI{
					ID:          uuidOrFail(t),
					URI:         fmt.Sprintf("https://test-%d-%d.impractical.services/testing", x, y),
					IsBaseURI:   (x+y)%2 == 0,
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

			res, err := storer.ListRedirectURIs(ctx, client.ID)
			if err != nil {
				t.Errorf("Error retrieving redirect URIs: %s", err)
			}
			clients.RedirectURIsByURI(redirectURIs)
			if len(redirectURIs) != len(res) {
				t.Fatalf("Expected %d results, got %d", len(redirectURIs), len(res))
			}
			for pos, uri := range redirectURIs {
				ok, field, expected, got := compareRedirectURIs(uri, res[pos])
				if !ok {
					t.Errorf("Expected %v for %q, got %v", expected, field, got)
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
			ok, field, expected, got := compareRedirectURIs(uri, res[pos])
			if !ok {
				t.Errorf("Expected %v for %q, got %v", expected, field, got)
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

// TODO: test creating a redirect URI with an ID that already exists

// TODO: test creating a redirect URI with a URI that already exists

// TODO: test removing a redirect URI that doesn't exist
