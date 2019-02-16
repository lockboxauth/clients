package storers

import (
	"time"

	"impractical.co/auth/clients"
)

type postgresClient struct {
	ID           string    `sql_column:"id"`
	SecretHash   string    `sql_column:"secret_hash"`
	SecretScheme string    `sql_column:"secret_scheme"`
	Confidential bool      `sql_column:"confidential"`
	CreatedAt    time.Time `sql_column:"created_at"`
	CreatedBy    string    `sql_column:"created_by"`
	CreatedByIP  string    `sql_column:"created_by_ip"`
}

func (p postgresClient) GetSQLTableName() string {
	return "clients"
}

func fromPostgres(c postgresClient) clients.Client {
	return clients.Client{
		ID:           c.ID,
		SecretHash:   c.SecretHash,
		SecretScheme: c.SecretScheme,
		Confidential: c.Confidential,
		CreatedAt:    c.CreatedAt,
		CreatedBy:    c.CreatedBy,
		CreatedByIP:  c.CreatedByIP,
	}
}

func toPostgres(c clients.Client) postgresClient {
	return postgresClient{
		ID:           c.ID,
		SecretHash:   c.SecretHash,
		SecretScheme: c.SecretScheme,
		Confidential: c.Confidential,
		CreatedAt:    c.CreatedAt,
		CreatedBy:    c.CreatedBy,
		CreatedByIP:  c.CreatedByIP,
	}
}

type postgresRedirectURI struct {
	ID          string    `sql_column:"id"`
	URI         string    `sql_column:"uri"`
	IsBaseURI   bool      `sql_column:"is_base_uri"`
	ClientID    string    `sql_column:"client_id"`
	CreatedAt   time.Time `sql_column:"created_at"`
	CreatedBy   string    `sql_column:"created_by"`
	CreatedByIP string    `sql_column:"created_by_ip"`
}

func (p postgresRedirectURI) GetSQLTableName() string {
	return "redirect_uris"
}

func uriFromPostgres(u postgresRedirectURI) clients.RedirectURI {
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

func uriToPostgres(u clients.RedirectURI) postgresRedirectURI {
	return postgresRedirectURI{
		ID:          u.ID,
		URI:         u.URI,
		IsBaseURI:   u.IsBaseURI,
		ClientID:    u.ClientID,
		CreatedAt:   u.CreatedAt,
		CreatedBy:   u.CreatedBy,
		CreatedByIP: u.CreatedByIP,
	}
}
