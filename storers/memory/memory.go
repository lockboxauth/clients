package memory

import (
	"context"
	"fmt"

	memdb "github.com/hashicorp/go-memdb"

	"lockbox.dev/clients"
)

var (
	schema = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"client": {
				Name: "client",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID", Lowercase: true},
					},
				},
			},
			"redirect_uri": {
				Name: "redirect_uri",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID", Lowercase: true},
					},
					"uri": {
						Name:    "uri",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "URI"},
					},
					"client_id": {
						Name:    "client_id",
						Indexer: &memdb.StringFieldIndex{Field: "ClientID"},
					},
				},
			},
		},
	}
)

// Storer is an in-memory implementation of the Storer
// interface.
type Storer struct {
	db *memdb.MemDB
}

// NewStorer returns a Storer instance that is ready
// to be used as a Storer.
func NewStorer() (*Storer, error) {
	db, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, err
	}
	return &Storer{
		db: db,
	}, nil
}

// Create inserts the passed clients.Client into the in-memory database. If
// another client in the in-memory database has the same value for its ID
// property, a clients.ErrClientAlreadyExists error is returned.
func (s Storer) Create(_ context.Context, client clients.Client) error {
	txn := s.db.Txn(true)
	defer txn.Abort()
	exists, err := txn.First("client", "id", client.ID)
	if err != nil {
		return err
	}
	if exists != nil {
		return clients.ErrClientAlreadyExists
	}
	err = txn.Insert("client", &client)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

// Get retrieves a clients.Client from the in-memory database if it can find
// one with an ID property matching the passed id. If a clients.Client with an
// ID property that matches the passed id can't be found, a
// clients.ErrClientNotFound is returned.
func (s Storer) Get(_ context.Context, id string) (clients.Client, error) {
	txn := s.db.Txn(false)
	defer txn.Abort()
	client, err := txn.First("client", "id", id)
	if err != nil {
		return clients.Client{}, err
	}
	if client == nil {
		return clients.Client{}, clients.ErrClientNotFound
	}
	res, ok := client.(*clients.Client)
	if !ok || res == nil {
		return clients.Client{}, fmt.Errorf("unexpected result type %T, wanted %T", client, new(clients.Client)) //nolint:goerr113 // this is just a test-facing error
	}
	return *res, nil
}

// Update apples the suppled clients.Change to any clients.Client in the
// in-memory database that has an ID property matching the passed id. If no
// clients.Client in the database has an ID property matching the passed id, no
// error is returned.
func (s Storer) Update(_ context.Context, id string, change clients.Change) error {
	txn := s.db.Txn(true)
	defer txn.Abort()
	client, err := txn.First("client", "id", id)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}
	res, ok := client.(*clients.Client)
	if !ok || res == nil {
		return fmt.Errorf("unexpected response type %T, expected %T", res, new(clients.Client)) //nolint:goerr113 // there is no recovering from this
	}
	updated := clients.Apply(change, *res)
	err = txn.Insert("client", &updated)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

// Delete removes any clients.Client in the in-memory database that has an ID
// property that matches the passed id. If no clients.Client in the database
// has an ID property that matches the passed id, no error is returned.
func (s Storer) Delete(_ context.Context, id string) error {
	txn := s.db.Txn(true)
	defer txn.Abort()
	exists, err := txn.First("client", "id", id)
	if err != nil {
		return err
	}
	if exists == nil {
		return nil
	}
	err = txn.Delete("client", exists)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

// ListRedirectURIs returns a []clients.RedirectURI containing all the
// clients.RedirectURIs in the in-memory database that have a ClientID property
// that matches clientID. If no clients.RedirectURIs in the database have a
// ClientID property that matches the passed clientID, an empty slice and nil
// error are returned. The slice is always sorted lexicographically by the URI.
func (s Storer) ListRedirectURIs(_ context.Context, clientID string) ([]clients.RedirectURI, error) {
	txn := s.db.Txn(false)
	var uris []clients.RedirectURI
	uriIter, err := txn.Get("redirect_uri", "client_id", clientID)
	if err != nil {
		return nil, err
	}
	for {
		uri := uriIter.Next()
		if uri == nil {
			break
		}
		redirURI, ok := uri.(*clients.RedirectURI)
		if !ok || redirURI == nil {
			return nil, fmt.Errorf("unexpected response type %T, expected %T", uri, new(clients.RedirectURI)) //nolint:goerr113 // there is no recovering from this
		}
		uris = append(uris, *redirURI)
	}
	clients.RedirectURIsByURI(uris)
	return uris, nil
}

// AddRedirectURIs persists the supplied clients.RedirectURIs in the in-memory
// database. If a clients.RedirectURI already exists in the database that has
// the same ID property as one of the specified clients.RedirectURIs, a
// clients.RedirectURIAlreadyExistsError will be returned with the ID property
// set. If a clients.RedirectURI already exists in the database that has the
// same URI as one of the specified clients.RedirectURIs, a
// clients.RedirectURIAlreadyExistsError will be returned with the URI property
// set. No validation is done that the ClientID property of the passed
// clients.RedirectURIs refers to a clients.Client in the database.
func (s Storer) AddRedirectURIs(_ context.Context, uris []clients.RedirectURI) error {
	txn := s.db.Txn(true)
	defer txn.Abort()
	for _, uri := range uris {
		exists, err := txn.First("redirect_uri", "id", uri.ID)
		if err != nil {
			return err
		}
		if exists != nil {
			return clients.RedirectURIAlreadyExistsError{ID: uri.ID}
		}
		exists, err = txn.First("redirect_uri", "uri", uri.URI)
		if err != nil {
			return err
		}
		if exists != nil {
			return clients.RedirectURIAlreadyExistsError{URI: uri.URI}
		}
		u := uri
		err = txn.Insert("redirect_uri", &u)
		if err != nil {
			return err
		}
	}
	txn.Commit()
	return nil
}

// RemoveRedirectURIs deletes any clients.RedirectURI in the in-memory database
// that has an ID property matching one of the passed ids. No error is returned
// if a passed id doesn't match to a clients.RedirectURI in the database.
func (s Storer) RemoveRedirectURIs(_ context.Context, ids []string) error {
	txn := s.db.Txn(true)
	defer txn.Abort()
	for _, id := range ids {
		exists, err := txn.First("redirect_uri", "id", id)
		if err != nil {
			return err
		}
		if exists == nil {
			continue
		}
		err = txn.Delete("redirect_uri", exists)
		if err != nil {
			return err
		}
	}
	txn.Commit()
	return nil
}
