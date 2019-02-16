package clients

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"time"
)

const (
	secretSchemeSHA256 = "sha256" // #nosec
)

var (
	// ErrClientAlreadyExists is returned when a client with the same ID
	// already exists in a Storer.
	ErrClientAlreadyExists = errors.New("client already exists")
	// ErrClientNotFound is returned when a client can't be located in a
	// Storer.
	ErrClientNotFound = errors.New("client not found")
	// ErrIncorrectSecret is returned when a client tries to authenticate
	// with an invalid secret.
	ErrIncorrectSecret = errors.New("incorrect client secret")
	// ErrUnsupportedSecretScheme is returned when a client uses a secret
	// scheme that we don't know how to use.
	ErrUnsupportedSecretScheme = errors.New("an unsupported secret scheme was used")
)

/// go:generate go-bindata -pkg migrations -o migrations/generated.go sql/

// Client represents an API client.
type Client struct {
	ID           string    // unique ID per client
	SecretHash   string    // hash of unique secret to authenticate with (optional)
	SecretScheme string    // the hashing scheme used for the secret
	Confidential bool      // whether this is a confidential (true) or public (false) client
	CreatedAt    time.Time // timestamp of creation
	CreatedBy    string    // the HMAC key that created this client
	CreatedByIP  string    // the IP that created this client
}

// CheckSecret returns nil if the passed secret is correct for the Client, or
// ErrIncorrectSecret if the secret is incorrect. Any other error signals data
// corruption.
func (c Client) CheckSecret(attempt string) error {
	switch c.SecretScheme {
	case secretSchemeSHA256:
		hashed, err := hex.DecodeString(c.SecretHash)
		if err != nil {
			return err
		}
		candidate := sha256.New().Sum([]byte(attempt))
		length := len(hashed)
		if len(candidate) > length {
			length = len(candidate)
		}
		consistentCandidate := make([]byte, length)
		consistentExpected := make([]byte, length)
		subtle.ConstantTimeCopy(1, consistentCandidate, candidate)
		subtle.ConstantTimeCopy(1, consistentExpected, hashed)
		if subtle.ConstantTimeCompare(consistentCandidate, consistentExpected) != 1 {
			return ErrIncorrectSecret
		}
	default:
		return ErrUnsupportedSecretScheme
	}
	return nil
}

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

// ErrRedirectURIAlreadyExists is returned when a redirect URI already exists
// in a Storer.
type ErrRedirectURIAlreadyExists struct {
	ID  string
	URI string // the URI that already exists
}

func (e ErrRedirectURIAlreadyExists) Error() string {
	if e.ID == "" {
		return fmt.Sprintf("redirect URI %q already exists", e.URI)
	}
	return fmt.Sprintf("redirect URI %q already exists", e.ID)
}

// Change represents a change we'd like to make to a Client. Nil values always
// represent "no change", whereas empty values will be interpreted as a desire
// to set the property to the empty value.
type Change struct {
	SecretHash   *string
	SecretScheme *string
}

// IsEmpty returns true if none of the fields in Change are set.
func (c Change) IsEmpty() bool {
	if c.SecretHash != nil {
		return false
	}
	if c.SecretScheme != nil {
		return false
	}
	return true
}

// ChangeSecret generates a Change that will update a Client's secret.
func ChangeSecret(newSecret []byte) (Change, error) {
	secret := hex.EncodeToString(sha256.New().Sum(newSecret))
	scheme := secretSchemeSHA256
	return Change{
		SecretHash:   &secret,
		SecretScheme: &scheme,
	}, nil
}

// Apply returns a Client with Change applied to it.
func Apply(change Change, client Client) Client {
	if change.IsEmpty() {
		return client
	}
	res := client
	if change.SecretHash != nil {
		res.SecretHash = *change.SecretHash
	}
	if change.SecretScheme != nil {
		res.SecretScheme = *change.SecretScheme
	}
	return res
}

// Storer is an interface for storing, retrieving, and modifying Clients and
// the metadata surrounding them.
type Storer interface {
	Create(ctx context.Context, client Client) error
	Get(ctx context.Context, id string) (Client, error)
	ListRedirectURIs(ctx context.Context, clientID string) ([]RedirectURI, error)
	Update(ctx context.Context, id string, change Change) error
	Delete(ctx context.Context, id string) error
	AddRedirectURIs(ctx context.Context, clientID string, uris []RedirectURI) error
	RemoveRedirectURIs(ctx context.Context, clientID string, uris []string) error
}

func RedirectURIsByURI(uris []RedirectURI) {
	sort.Slice(uris, func(i, j int) bool {
		return uris[i].URI < uris[j].URI
	})
}
