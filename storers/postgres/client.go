package postgres

import (
	"time"

	"lockbox.dev/clients"
)

// Client is a representation of the clients.Client type that is suitable to be
// stored in a PostgreSQL database.
type Client struct {
	ID           string    `sql_column:"id"`
	Name         string    `sql_column:"name"`
	SecretHash   string    `sql_column:"secret_hash"`
	SecretScheme string    `sql_column:"secret_scheme"`
	Confidential bool      `sql_column:"confidential"`
	CreatedAt    time.Time `sql_column:"created_at"`
	CreatedBy    string    `sql_column:"created_by"`
	CreatedByIP  string    `sql_column:"created_by_ip"`
}

// GetSQLTableName returns the name of the SQL table that the data for this
// type will be stored in.
func (Client) GetSQLTableName() string {
	return "clients"
}

func fromPostgres(client Client) clients.Client {
	return clients.Client{
		ID:           client.ID,
		Name:         client.Name,
		SecretHash:   client.SecretHash,
		SecretScheme: client.SecretScheme,
		Confidential: client.Confidential,
		CreatedAt:    client.CreatedAt,
		CreatedBy:    client.CreatedBy,
		CreatedByIP:  client.CreatedByIP,
	}
}

func toPostgres(client clients.Client) Client {
	return Client{
		ID:           client.ID,
		Name:         client.Name,
		SecretHash:   client.SecretHash,
		SecretScheme: client.SecretScheme,
		Confidential: client.Confidential,
		CreatedAt:    client.CreatedAt,
		CreatedBy:    client.CreatedBy,
		CreatedByIP:  client.CreatedByIP,
	}
}
