// Package clients provides a record of conceptual clients that will be
// interacting with Lockbox.
//
// Clients represent an actor in the system, a deployment of software that
// needs access to certain Lockbox (and third-party) APIs. Clients are largely
// useful for keeping track of where requests came from and limiting the scopes
// available in certain situations.
//
// The clients package provides the definitions of the service and its
// boundaries. It sets up the Client type, which represents an API consumer,
// the RedirectURI type, which represents a URI that a client's authentication
// requests are able to be redirected to, and the Storer interface, which
// defines how to implement data storage backends for these Clients and
// RedirectURIs.
//
// This package can be thought of as providing the types and helpers that form
// the conceptual framework of the subsystem, but with very little
// functionality provided by itself. Instead, implementations of the interfaces
// and sub-packages using these types are where most functionality will
// actually live.
package clients
