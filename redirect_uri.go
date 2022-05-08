package clients

import (
	"fmt"
	"sort"
	"time"
)

// RedirectURI represents a URI that we'll redirect to as part of the OAuth 2
// dance for a Client. The RedirectURI is an important part of authorizing a
// client, especially a public one, as it prevents others from using a Client's
// ID.
type RedirectURI struct {
	ID          string    // unique ID per redirect URI
	URI         string    // the URI to redirect to
	IsBaseURI   bool      // whether this is the full URI (false) or just a base (true)
	ClientID    string    // the ID of the Client this redirect URI applies to
	CreatedAt   time.Time // the timestamp this redirect URI was created at
	CreatedBy   string    // the HMAC key that created this redirect URI
	CreatedByIP string    // the IP that created this redirect URI
}

// RedirectURIAlreadyExistsError is returned when a redirect URI already exists
// in a Storer.
type RedirectURIAlreadyExistsError struct {
	ID  string
	URI string // the URI that already exists
	Err error  // the error that was returned, if any
}

// Error fills the error interface for RedirectURIs.
func (e RedirectURIAlreadyExistsError) Error() string {
	if e.ID == "" && e.URI == "" && e.Err != nil {
		return e.Err.Error()
	}
	if e.ID == "" {
		return fmt.Sprintf("redirect URI %q already exists", e.URI)
	}
	return fmt.Sprintf("redirect URI %q already exists", e.ID)
}

// RedirectURIsByURI returns `uris` sorted by their URI property, with
// URIs that are lexicographically lower returned first.
func RedirectURIsByURI(uris []RedirectURI) {
	sort.Slice(uris, func(i, j int) bool {
		return uris[i].URI < uris[j].URI
	})
}
