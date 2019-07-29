package memory

import (
	"context"

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

func (s Storer) Create(ctx context.Context, client clients.Client) error {
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

func (s Storer) Get(ctx context.Context, id string) (clients.Client, error) {
	txn := s.db.Txn(false)
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

func (s Storer) Update(ctx context.Context, id string, change clients.Change) error {
	txn := s.db.Txn(true)
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

func (s Storer) Delete(ctx context.Context, id string) error {
	txn := s.db.Txn(true)
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

func (s Storer) ListRedirectURIs(ctx context.Context, clientID string) ([]clients.RedirectURI, error) {
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
		uris = append(uris, *uri.(*clients.RedirectURI))
	}
	clients.RedirectURIsByURI(uris)
	return uris, nil
}

func (s Storer) AddRedirectURIs(ctx context.Context, uris []clients.RedirectURI) error {
	txn := s.db.Txn(true)
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

func (s Storer) RemoveRedirectURIs(ctx context.Context, uris []string) error {
	txn := s.db.Txn(true)
	defer txn.Abort()
	for _, uri := range uris {
		exists, err := txn.First("redirect_uri", "id", uri)
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
