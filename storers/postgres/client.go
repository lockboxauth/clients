package postgres

import (
	"time"

	"impractical.co/auth/clients"
)

type Client struct {
	ID           string    `sql_column:"id"`
	SecretHash   string    `sql_column:"secret_hash"`
	SecretScheme string    `sql_column:"secret_scheme"`
	Confidential bool      `sql_column:"confidential"`
	CreatedAt    time.Time `sql_column:"created_at"`
	CreatedBy    string    `sql_column:"created_by"`
	CreatedByIP  string    `sql_column:"created_by_ip"`
}

func (p Client) GetSQLTableName() string {
	return "clients"
}

func fromPostgres(c Client) clients.Client {
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

func toPostgres(c clients.Client) Client {
	return Client{
		ID:           c.ID,
		SecretHash:   c.SecretHash,
		SecretScheme: c.SecretScheme,
		Confidential: c.Confidential,
		CreatedAt:    c.CreatedAt,
		CreatedBy:    c.CreatedBy,
		CreatedByIP:  c.CreatedByIP,
	}
}
