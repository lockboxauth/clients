package postgres

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strings"

	"darlinggo.co/pan"
	"yall.in"

	"lockbox.dev/clients"

	"github.com/lib/pq"
)

var (
	redirectURIValueRegex = regexp.MustCompile(`^Key \(([^)]*)\)=\(([^)]*)\) already exists.$`)
)

const (
	redirectURIValueRegexGroups = 3
)

//go:generate go-bindata -pkg migrations -o migrations/generated.go sql/

// Storer is an implementation of the Storer interface that stores data in a
// PostgreSQL database.
type Storer struct {
	db *sql.DB
}

// NewStorer returns a Storer instance that is backed by the specified *sql.DB.
// The returned Storer instance is ready to be used as a clients.Storer.
func NewStorer(_ context.Context, conn *sql.DB) *Storer {
	return &Storer{db: conn}
}

// Create inserts the passed clients.Client into the database, returning an
// error if it cannot. If the clients.Client already exists in the database, a
// clients.ErrClientAlreadyExists error is returned.
func (s Storer) Create(ctx context.Context, client clients.Client) error {
	query := createSQL(ctx, toPostgres(client))
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(queryStr, query.Args()...)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Constraint == "clients_pkey" {
		err = clients.ErrClientAlreadyExists
	}
	return err
}

// Get retrieves the clients.Client in the database with an id column that
// matches the passed id. If one can't be found, a clients.ErrClientNotFound
// error is returned.
func (s Storer) Get(ctx context.Context, id string) (clients.Client, error) {
	query := getSQL(ctx, id)
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return clients.Client{}, err
	}
	rows, err := s.db.QueryContext(ctx, queryStr, query.Args()...) //nolint:sqlclosecheck // the closeRows helper isn't picked up
	if err != nil {
		return clients.Client{}, err
	}
	defer closeRows(ctx, rows)
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

// ListRedirectURIs finds all the clients.RedirectURIs in the PostgreSQL
// database that have a client_id column that matches the passed clientID. If
// there are none, an empty slice and a nil error are returned.
func (s Storer) ListRedirectURIs(ctx context.Context, clientID string) ([]clients.RedirectURI, error) {
	query := listRedirectURIsSQL(ctx, clientID)
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, queryStr, query.Args()...) //nolint:sqlclosecheck // it's closed, it's just not picking up the closeRows helper
	if err != nil {
		return nil, err
	}
	defer closeRows(ctx, rows)
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

// Update applies the passed clients.Change to the clients.Client in the
// database with an id column matching the passed id. If row matches, no error
// is returned.
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

// Delete removes any rows with an id column matching the passed id from the
// clients table in the database. If no rows match, no error is returned.
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

// AddRedirectURIs inserts a group of clients.RedirectURIs into the database.
// The clients.RedirectURIs do not need to be for the same clients.Client, and
// no validation is done that the clients.RedirectURIs are being associated
// with a clients.Client that exists. If the ID of any clients.RedirectURI is
// already in the database, a clients.RedirectURIAlreadyExistsError with the ID
// property set is returned. If the URI of any clients.RedirectURI is already
// in the database, a clients.RedirectURIAlreadyExistsError is returned with
// the URI property set.
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
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return err
	}
	redErr := clients.RedirectURIAlreadyExistsError{
		Err: pqErr,
	}
	switch pqErr.Constraint {
	case "redirect_uris_pkey":
		matches := redirectURIValueRegex.FindStringSubmatch(pqErr.Detail)
		if len(matches) < redirectURIValueRegexGroups {
			yall.FromContext(ctx).WithError(err).WithField("matches", len(matches)).Error("unexpected number of redirect URI constraint error matches")
			return redErr
		}
		if matches[1] != "id" {
			yall.FromContext(ctx).WithError(err).WithField("column", matches[1]).Error("unexpected column for redirect URI constraint error")
			return redErr
		}
		redErr.ID = strings.TrimSpace(matches[2])
	case "redirect_uris_unique_uri":
		matches := redirectURIValueRegex.FindStringSubmatch(pqErr.Detail)
		if len(matches) < redirectURIValueRegexGroups {
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

// RemoveRedirectURIs deletes the redirect URIs with the passed IDs from the
// database. If an ID is not found, it is ignored.
func (s Storer) RemoveRedirectURIs(ctx context.Context, ids []string) error {
	query := removeRedirectURIsSQL(ctx, ids)
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

func closeRows(ctx context.Context, rows *sql.Rows) {
	if err := rows.Close(); err != nil {
		yall.FromContext(ctx).WithError(err).Error("failed to close rows")
	}
}
