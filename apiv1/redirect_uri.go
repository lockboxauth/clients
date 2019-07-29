package apiv1

import (
	"time"

	"lockbox.dev/clients"
)

// RedirectURI is an API-specific representation of a redirect URI.
type RedirectURI struct {
	ID          string    `json:"ID"`
	URI         string    `json:"URI"`
	IsBaseURI   bool      `json:"isBaseURI"`
	ClientID    string    `json:"clientID"`
	CreatedAt   time.Time `json:"createdAt"`
	CreatedBy   string    `json:"createdBy"`
	CreatedByIP string    `json:"createdByIP"`
}

func coreRedirectURI(redirectURI RedirectURI) clients.RedirectURI {
	return clients.RedirectURI{
		ID:          redirectURI.ID,
		URI:         redirectURI.URI,
		IsBaseURI:   redirectURI.IsBaseURI,
		ClientID:    redirectURI.ClientID,
		CreatedAt:   redirectURI.CreatedAt,
		CreatedBy:   redirectURI.CreatedBy,
		CreatedByIP: redirectURI.CreatedByIP,
	}
}

func coreRedirectURIs(uris []RedirectURI) []clients.RedirectURI {
	res := make([]clients.RedirectURI, 0, len(uris))
	for _, uri := range uris {
		res = append(res, coreRedirectURI(uri))
	}
	return res
}

func apiRedirectURI(redirectURI clients.RedirectURI) RedirectURI {
	return RedirectURI{
		ID:          redirectURI.ID,
		URI:         redirectURI.URI,
		IsBaseURI:   redirectURI.IsBaseURI,
		ClientID:    redirectURI.ClientID,
		CreatedAt:   redirectURI.CreatedAt,
		CreatedBy:   redirectURI.CreatedBy,
		CreatedByIP: redirectURI.CreatedByIP,
	}
}

func apiRedirectURIs(uris []clients.RedirectURI) []RedirectURI {
	res := make([]RedirectURI, 0, len(uris))
	for _, uri := range uris {
		res = append(res, apiRedirectURI(uri))
	}
	return res
}
