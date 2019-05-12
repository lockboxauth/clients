package clients

import "context"

// Storer is an interface for storing, retrieving, and modifying Clients and
// the metadata surrounding them.
type Storer interface {
	Create(ctx context.Context, client Client) error
	Get(ctx context.Context, id string) (Client, error)
	ListRedirectURIs(ctx context.Context, clientID string) ([]RedirectURI, error)
	Update(ctx context.Context, id string, change Change) error
	Delete(ctx context.Context, id string) error
	AddRedirectURIs(ctx context.Context, uris []RedirectURI) error
	RemoveRedirectURIs(ctx context.Context, uris []string) error
}
