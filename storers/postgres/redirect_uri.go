package postgres

import (
	"time"

	"lockbox.dev/clients"
)

// RedirectURI is a representation of the clients.RedirectURI type that is
// suitable to be stored in a PostgreSQL database.
type RedirectURI struct {
	ID          string    `sql_column:"id"`
	URI         string    `sql_column:"uri"`
	IsBaseURI   bool      `sql_column:"is_base_uri"`
	ClientID    string    `sql_column:"client_id"`
	CreatedAt   time.Time `sql_column:"created_at"`
	CreatedBy   string    `sql_column:"created_by"`
	CreatedByIP string    `sql_column:"created_by_ip"`
}

// GetSQLTableName returns the name of the SQL table that the data for this
// type will be stored in.
func (RedirectURI) GetSQLTableName() string {
	return "redirect_uris"
}

func uriFromPostgres(uri RedirectURI) clients.RedirectURI {
	return clients.RedirectURI{
		ID:          uri.ID,
		URI:         uri.URI,
		IsBaseURI:   uri.IsBaseURI,
		ClientID:    uri.ClientID,
		CreatedAt:   uri.CreatedAt,
		CreatedBy:   uri.CreatedBy,
		CreatedByIP: uri.CreatedByIP,
	}
}

func uriToPostgres(uri clients.RedirectURI) RedirectURI {
	return RedirectURI{
		ID:          uri.ID,
		URI:         uri.URI,
		IsBaseURI:   uri.IsBaseURI,
		ClientID:    uri.ClientID,
		CreatedAt:   uri.CreatedAt,
		CreatedBy:   uri.CreatedBy,
		CreatedByIP: uri.CreatedByIP,
	}
}
