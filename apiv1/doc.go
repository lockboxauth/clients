// Package apiv1 provides a JSON API for interacting with clients.
//
// This package can be imported to get an http.Handler that will enable
// creating, reading, updating, and deleting Clients.
//
// The lockbox.dev/hmac package is used to authenticate requests with HMAC
// authentication. Authentication grants read and write access for all clients
// and redirect URIs in the clients system. Authentication is meant to be
// simple and to distinguish administrators from unauthorized users; no other
// roles are expected to interact with the API. The Key used to sign a request
// will be stored as the CreatedBy property on clients and redirect URIs. The
// IP the request was made from will be stored as CreatedByIP.
package apiv1
