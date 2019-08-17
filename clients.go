package clients

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
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

// Client represents an API client.
type Client struct {
	ID           string    // unique ID per client
	Name         string    // friendly name for this client
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

// Change represents a change we'd like to make to a Client. Nil values always
// represent "no change", whereas empty values will be interpreted as a desire
// to set the property to the empty value.
type Change struct {
	Name         *string
	SecretHash   *string
	SecretScheme *string
}

// IsEmpty returns true if none of the fields in Change are set.
func (c Change) IsEmpty() bool {
	if c.Name != nil {
		return false
	}
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
	if change.Name != nil {
		res.Name = *change.Name
	}
	if change.SecretHash != nil {
		res.SecretHash = *change.SecretHash
	}
	if change.SecretScheme != nil {
		res.SecretScheme = *change.SecretScheme
	}
	return res
}
