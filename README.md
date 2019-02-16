# clients

The `clients` package encapsulates the part of the [auth system](https://impractical.co/auth) that defines API clients and provides management interfaces for them.

## Scope

`clients` is solely responsible for managing the list of clients and their authentication information, along with any restrictions placed on which scopes they may use.

The questions `clients` is meant to answer for the system include:

  * Is this a valid API client?
  * Which scopes can this client access?
  * Is this a valid URI to redirect to for this client?
  * How should this client be displayed to the user?
