package postgres

import (
	"context"
	"database/sql"
	"regexp"
	"strings"

	"darlinggo.co/pan"
	"yall.in"

	"lockbox.dev/clients"

	"github.com/lib/pq"
)

var redirectURIValueRegex = regexp.MustCompile(`^Key \(([^)]*)\)=\(([^)]*)\) already exists.$`)

//go:generate go-bindata -pkg migrations -o migrations/generated.go sql/

// Postgres is an implementation of the Storer interface
// that stores data in a PostgreSQL database.
type Storer struct {
	db *sql.DB
}

// NewPostgres returns a Postgres instance that is backed by the specified
// *sql.DB. The returned Postgres instance is ready to be used as a Storer.
func NewStorer(ctx context.Context, conn *sql.DB) *Storer {
	return &Storer{db: conn}
}

func (s Storer) Create(ctx context.Context, client clients.Client) error {
	query := createSQL(ctx, toPostgres(client))
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(queryStr, query.Args()...)
	if e, ok := err.(*pq.Error); ok {
		if e.Constraint == "clients_pkey" {
			err = clients.ErrClientAlreadyExists
		}
	}
	return err
}

func (s Storer) Get(ctx context.Context, id string) (clients.Client, error) {
	query := getSQL(ctx, id)
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return clients.Client{}, err
	}
	rows, err := s.db.Query(queryStr, query.Args()...)
	if err != nil {
		return clients.Client{}, err
	}
	var client Client
	for rows.Next() {
		err = pan.Unmarshal(rows, &client)
		if err != nil {
			return clients.Client{}, err
		}
	}
	if err = rows.Err(); err != nil {
		return clients.Client{}, err
	}
	if client.ID == "" {
		return clients.Client{}, clients.ErrClientNotFound
	}
	return fromPostgres(client), nil
}

func (s Storer) ListRedirectURIs(ctx context.Context, clientID string) ([]clients.RedirectURI, error) {
	query := listRedirectURIsSQL(ctx, clientID)
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return nil, err
	}
	rows, err := s.db.Query(queryStr, query.Args()...)
	if err != nil {
		return nil, err
	}
	var results []clients.RedirectURI
	for rows.Next() {
		var uri RedirectURI
		err = pan.Unmarshal(rows, &uri)
		if err != nil {
			return results, err
		}
		results = append(results, uriFromPostgres(uri))
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	clients.RedirectURIsByURI(results)
	return results, nil
}

func (s Storer) Update(ctx context.Context, id string, change clients.Change) error {
	if change.IsEmpty() {
		return nil
	}
	query := updateSQL(ctx, id, change)
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(queryStr, query.Args()...)
	if err != nil {
		return err
	}
	return nil
}

func (s Storer) Delete(ctx context.Context, id string) error {
	query := deleteSQL(ctx, id)
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(queryStr, query.Args()...)
	if err != nil {
		return err
	}
	return nil
}

func (s Storer) AddRedirectURIs(ctx context.Context, uris []clients.RedirectURI) error {
	pgURIs := make([]RedirectURI, 0, len(uris))
	for _, uri := range uris {
		pgURIs = append(pgURIs, uriToPostgres(uri))
	}
	query := addRedirectURIsSQL(ctx, pgURIs)
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(queryStr, query.Args()...)
	if e, ok := err.(*pq.Error); ok {
		redErr := clients.ErrRedirectURIAlreadyExists{
			Err: e,
		}
		if e.Constraint == "redirect_uris_pkey" {
			matches := redirectURIValueRegex.FindStringSubmatch(e.Detail)
			if len(matches) < 3 {
				yall.FromContext(ctx).WithError(err).WithField("matches", len(matches)).Error("unexpected number of redirect URI constraint error matches")
				return redErr
			}
			if matches[1] != "id" {
				yall.FromContext(ctx).WithError(err).WithField("column", matches[1]).Error("unexpected column for redirect URI constraint error")
				return redErr
			}
			redErr.ID = strings.TrimSpace(matches[2])
		} else if e.Constraint == "redirect_uris_unique_uri" {
			matches := redirectURIValueRegex.FindStringSubmatch(e.Detail)
			if len(matches) < 3 {
				yall.FromContext(ctx).WithError(err).WithField("matches", len(matches)).Error("unexpected number of redirect URI constraint error matches")
				return redErr
			}
			if matches[1] != "uri" {
				yall.FromContext(ctx).WithError(err).WithField("matches", len(matches)).Error("unexpected column for redirect URI constraint error")
				return redErr
			}
			redErr.URI = strings.TrimSpace(matches[2])
		}
		return redErr
	}
	return err
}

func (s Storer) RemoveRedirectURIs(ctx context.Context, uris []string) error {
	query := removeRedirectURIsSQL(ctx, uris)
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(queryStr, query.Args()...)
	if err != nil {
		return err
	}
	return nil
}
