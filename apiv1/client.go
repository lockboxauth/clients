package apiv1

import (
	"time"

	"lockbox.dev/clients"
)

// Client is an API-specific representation of a client.
type Client struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Confidential bool      `json:"confidential"`
	CreatedAt    time.Time `json:"createdAt"`
	CreatedBy    string    `json:"createdBy"`
	CreatedByIP  string    `json:"createdByIP"`
	Secret       string    `json:"secret,omitempty"`
}

func coreClient(client Client) clients.Client {
	return clients.Client{
		ID:           client.ID,
		Name:         client.Name,
		Confidential: client.Confidential,
		CreatedAt:    client.CreatedAt,
		CreatedBy:    client.CreatedBy,
		CreatedByIP:  client.CreatedByIP,
	}
}

func apiClient(client clients.Client) Client {
	return Client{
		ID:           client.ID,
		Name:         client.Name,
		Confidential: client.Confidential,
		CreatedAt:    client.CreatedAt,
		CreatedBy:    client.CreatedBy,
		CreatedByIP:  client.CreatedByIP,
	}
}
