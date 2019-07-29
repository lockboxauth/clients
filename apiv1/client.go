package apiv1

import (
	"time"

	"lockbox.dev/clients"
)

// Client is an API-specific representation of a client.
type Client struct {
	ID           string    `json:"id"`
	Confidential bool      `json:"confidential"`
	CreatedAt    time.Time `json:"createdAt"`
	CreatedBy    string    `json:"createdBy"`
	CreatedByIP  string    `json:"createdByIP"`
	Secret       string    `json:"secret,omitempty"`
}

func coreClient(client Client) clients.Client {
	return clients.Client{
		ID:           client.ID,
		Confidential: client.Confidential,
		CreatedAt:    client.CreatedAt,
		CreatedBy:    client.CreatedBy,
		CreatedByIP:  client.CreatedByIP,
	}
}

func coreClients(cs []Client) []clients.Client {
	res := make([]clients.Client, 0, len(cs))
	for _, c := range cs {
		res = append(res, coreClient(c))
	}
	return res
}

func apiClient(client clients.Client) Client {
	return Client{
		ID:           client.ID,
		Confidential: client.Confidential,
		CreatedAt:    client.CreatedAt,
		CreatedBy:    client.CreatedBy,
		CreatedByIP:  client.CreatedByIP,
	}
}

func apiClients(cs []clients.Client) []Client {
	res := make([]Client, 0, len(cs))
	for _, c := range cs {
		res = append(res, apiClient(c))
	}
	return res
}
