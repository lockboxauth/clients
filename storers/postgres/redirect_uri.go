package postgres

import (
	"time"

	"impractical.co/auth/clients"
)

type RedirectURI struct {
	ID          string    `sql_column:"id"`
	URI         string    `sql_column:"uri"`
	IsBaseURI   bool      `sql_column:"is_base_uri"`
	ClientID    string    `sql_column:"client_id"`
	CreatedAt   time.Time `sql_column:"created_at"`
	CreatedBy   string    `sql_column:"created_by"`
	CreatedByIP string    `sql_column:"created_by_ip"`
}

func (p RedirectURI) GetSQLTableName() string {
	return "redirect_uris"
}

func uriFromPostgres(u RedirectURI) clients.RedirectURI {
	return clients.RedirectURI{
		ID:          u.ID,
		URI:         u.URI,
		IsBaseURI:   u.IsBaseURI,
		ClientID:    u.ClientID,
		CreatedAt:   u.CreatedAt,
		CreatedBy:   u.CreatedBy,
		CreatedByIP: u.CreatedByIP,
	}
}

func uriToPostgres(u clients.RedirectURI) RedirectURI {
	return RedirectURI{
		ID:          u.ID,
		URI:         u.URI,
		IsBaseURI:   u.IsBaseURI,
		ClientID:    u.ClientID,
		CreatedAt:   u.CreatedAt,
		CreatedBy:   u.CreatedBy,
		CreatedByIP: u.CreatedByIP,
	}
}
