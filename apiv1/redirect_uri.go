package apiv1

import (
	"impractical.co/auth/clients"
)

// RedirectURI is an API-specific representation of a redirect URI.
type RedirectURI struct {
	// TODO: fill in API fields
}

func coreRedirectURI(redirectURI RedirectURI) clients.RedirectURI {
	return clients.RedirectURI{
		// TODO: map API fields to RedirectURI struct
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
		// TODO: map API fields to RedirectURI struct
	}
}

func apiRedirectURIs(uris []clients.RedirectURI) []RedirectURI {
	res := make([]RedirectURI, 0, len(uris))
	for _, uri := range uris {
		res = append(res, apiRedirectURI(uri))
	}
	return res
}
