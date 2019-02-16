package storers

import (
	"context"
	"database/sql"
	"errors"

	"impractical.co/auth/clients"
)

// Postgres is an implementation of the Storer interface
// that stores data in a PostgreSQL database.
type Postgres struct {
	db *sql.DB
}

// NewPostgres returns a Postgres instance that is backed by the specified
// *sql.DB. The returned Postgres instance is ready to be used as a Storer.
func NewPostgres(ctx context.Context, conn *sql.DB) *Postgres {
	return &Postgres{db: conn}
}

func (p Postgres) Create(ctx context.Context, client clients.Client) error {
	// TODO: implement creating a client
	return errors.New("not implemented")
}

func (p Postgres) Get(ctx context.Context, id string) (clients.Client, error) {
	// TODO: implement retrieving a client
	return clients.Client{}, errors.New("not implemented")
}

func (p Postgres) ListRedirectURIs(ctx context.Context, clientID string) ([]clients.RedirectURI, error) {
	// TODO: implement listing redirect URIs
	return nil, errors.New("not implemented")
}

func (p Postgres) Update(ctx context.Context, id string, change clients.Change) error {
	// TODO: implement updating a client
	return errors.New("not implemented")
}

func (p Postgres) Delete(ctx context.Context, id string) error {
	// TODO: implement deleting a client
	return errors.New("not implemented")
}

func (p Postgres) AddRedirectURIs(ctx context.Context, clientID string, uris []clients.RedirectURI) error {
	// TODO: implement adding redirect URIs to a client
	return errors.New("not implemented")
}

func (p Postgres) RemoveRedirectURIs(ctx context.Context, clientID string, uris []string) error {
	// TODO: implement removing redirect URIs from a client
	return errors.New("not implemented")
}
