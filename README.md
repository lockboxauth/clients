# clients

`clients` is a subsystem of [Lockbox](https://lockbox.dev). Its responsibility
is to manage the clients that will be interacting with the system.

Clients are just software executing requests. The `clients` subsystem provides
for clients that have redirect-based authentication schemes and for clients
that have authentication schemes requiring a secret. The job of the `clients`
subsystem is to store the data on these clients.

## Design Goals

`clients` is meant to be a discrete subsystem in the overall
[Lockbox](https://lockbox.dev) system. It tries to have clear boundaries of
responsibility and limit its responsibilities to only the things that it is
uniquely situated to do. Like all Lockbox subsystems, `clients` is meant to be
an interchangeable part of the system, easily replaceable. All of its
functionality should be exposed through its API, instead of relying on other
subsystems importing it directly.

`clients` assumes at this point that a trusted system is executing requests
against its APIs, and its API uses an HMAC-based scheme for authentication that
will not scale well to untrusted third parties executing requests against it.
If third parties can register clients independently, it is recommended that a
separate service handles these registrations, calling through to the `clients`
subsystem on the backend. 

## Implementation

`clients` is implemented largely as a datastore and access mechanism for an
`Client` and `RedirectURI` types. `Client` types have a unique ID per client,
which is how they should be programmatically referenced, and a name, which is
how they should be referenced to end users. When `Client` types are generated
with a secret component, that component is stored as a PBKDF2 hash using
SHA-256. The data model allows for the hashing schema to be changed without
invalidating prior secrets.

The API uses an HMAC authentication scheme, expecting the request to be signed
with a secret that only authorized parties have. The server uses the secret to
verify the signature. The design allows for keys to be rotated while still
being able to accept signatures from older keys as part of the transition.
Anyone holding the secret is able to create, update, and delete any `Client`.

The HMAC authentication system was chosen to limit the complexity of the
limitation and allow it to not rely on other subsystems for authentication of
requests. This yields a more limited authentication system for `clients` API
requests, but one that has fewer dependencies.

`RedirectURI` types are URIs that are registered on `Client` types as a
redirect target for authentication purposes. These types have an opaque,
randomly generated ID, which is how they should be identified programmatically.
These URIs can either be a full URI or a base URI that will serve as a prefix
and allow clients to be authenticated by redirecting to any URL the request
specifies that begins with that prefix.

`Client` types may have zero or more `RedirectURI` types associated with them.
Each `RedirectURI` type may only be associated with a single `Client`.

## Scope

`clients` is solely responsible for managing the list of clients and their
authentication information, along with any restrictions placed on which scopes
they may use.

The questions `clients` is meant to answer for the system include:

  * Is this a valid API client?
  * Does this client use this secret?
  * Which scopes can this client access?
  * Is this a valid URI to redirect to for this client?
  * How should this client be displayed to the user?

## Repository Structure

The base directory of the repository is used to set the logical framework and
shared types that will be used to talk about the subsystem. This largely means
defining types and interfaces.

The storers directory contains a collection of implementations of the `Storer`
interface, each in their own package. These packages should only have unit
tests, if any tests. The `Storer` acceptance tests in `storer_test.go` have
common acceptance testing for `Storer` implementations, and all `Storer`
implementations in the storers directory should register their tests there. If
the tests have setup requirements like databases or credentials, the tests
should only register themselves if these credentials are found.

The apiv1 directory contains the first version of the API interface. Breaking
changes should be published in a separate apiv2 package, so that both versions
of the API can be run simultaneously.
