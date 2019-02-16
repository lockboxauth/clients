package storers

import (
	"context"
	"fmt"

	memdb "github.com/hashicorp/go-memdb"
	"impractical.co/auth/clients"
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

// Memstore is an in-memory implementation of the Storer
// interface.
type Memstore struct {
	db *memdb.MemDB
}

// NewMemstore returns a Memstore instance that is ready
// to be used as a Storer.
func NewMemstore() (*Memstore, error) {
	db, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, err
	}
	return &Memstore{
		db: db,
	}, nil
}

func (m Memstore) Create(ctx context.Context, client clients.Client) error {
	txn := m.db.Txn(true)
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

func (m Memstore) Get(ctx context.Context, id string) (clients.Client, error) {
	txn := m.db.Txn(false)
	defer txn.Abort()
	client, err := txn.First("client", "id", id)
	if err != nil {
		return clients.Client{}, err
	}
	if client == nil {
		return clients.Client{}, clients.ErrClientNotFound
	}
	return *client.(*clients.Client), nil
}

func (m Memstore) Update(ctx context.Context, id string, change clients.Change) error {
	txn := m.db.Txn(true)
	defer txn.Abort()
	client, err := txn.First("client", "id", id)
	if err != nil {
		return err
	}
	if client == nil {
		return clients.ErrClientNotFound
	}
	updated := clients.Apply(change, *client.(*clients.Client))
	err = txn.Insert("client", &updated)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func (m Memstore) Delete(ctx context.Context, id string) error {
	txn := m.db.Txn(true)
	defer txn.Abort()
	exists, err := txn.First("client", "id", id)
	if err != nil {
		return err
	}
	if exists == nil {
		return clients.ErrClientNotFound
	}
	err = txn.Delete("client", exists)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func (m Memstore) ListRedirectURIs(ctx context.Context, clientID string) ([]clients.RedirectURI, error) {
	txn := m.db.Txn(false)
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
		uris = append(uris, *uri.(*clients.RedirectURI))
	}
	clients.RedirectURIsByURI(uris)
	return uris, nil
}

func (m Memstore) AddRedirectURIs(ctx context.Context, clientID string, uris []clients.RedirectURI) error {
	txn := m.db.Txn(true)
	defer txn.Abort()
	for _, uri := range uris {
		exists, err := txn.First("redirect_uri", "id", uri.ID)
		if err != nil {
			return err
		}
		if exists != nil {
			return clients.ErrRedirectURIAlreadyExists{ID: uri.ID}
		}
		exists, err = txn.First("redirect_uri", "uri", uri.URI)
		if err != nil {
			return err
		}
		if exists != nil {
			return clients.ErrRedirectURIAlreadyExists{URI: uri.URI}
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

func (m Memstore) RemoveRedirectURIs(ctx context.Context, clientID string, uris []string) error {
	txn := m.db.Txn(true)
	defer txn.Abort()
	for _, uri := range uris {
		exists, err := txn.First("redirect_uri", "id", uri)
		if err != nil {
			return err
		}
		if exists == nil {
			continue
		}
		if exists.(*clients.RedirectURI).ClientID != clientID {
			return fmt.Errorf("URI %q doesn't belong to client %q", uri, clientID)
		}
		err = txn.Delete("redirect_uri", exists)
		if err != nil {
			return err
		}
	}
	txn.Commit()
	return nil
}
