package apiv1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"darlinggo.co/api"
	"impractical.co/auth/clients"
	yall "yall.in"
)

func (a APIv1) handleCreateClient(w http.ResponseWriter, r *http.Request) {
	input, resp := a.VerifyRequest(r)
	if resp != nil {
		api.Encode(w, r, resp.Status, resp)
		return
	}
	var body struct {
		Client       Client        `json:"client"`
		RedirectURIs []RedirectURI `json:"redirectURIs"`
	}
	err := json.Unmarshal([]byte(input), &body)
	if err != nil {
		yall.FromContext(r.Context()).WithError(err).Debug("Error decoding request body")
		api.Encode(w, r, http.StatusBadRequest, Response{Errors: api.InvalidFormatError})
		return
	}

	var reqErrs []api.RequestError

	// if this isn't a confidential client, redirect URIs must be specified
	if !body.Client.Confidential && len(body.RedirectURIs) < 1 {
		reqErrs = append(reqErrs, api.RequestError{Field: "/client/confidential,/redirectURIs", Slug: api.RequestErrConflict})
	}
	// TODO: further validation, probably?
	// TODO: fill defaults?
	if len(reqErrs) > 0 {
		api.Encode(w, r, http.StatusBadRequest, reqErrs)
		return
	}
	client := coreClient(body.Client)
	redirectURIs := coreRedirectURIs(body.RedirectURIs)
	err = a.Storer.Create(r.Context(), client)
	if err != nil {
		if err == clients.ErrClientAlreadyExists {
			api.Encode(w, r, http.StatusBadRequest, Response{Errors: []api.RequestError{{Field: "/client/id", Slug: api.RequestErrConflict}}})
			return
		}
		yall.FromContext(r.Context()).WithError(err).Error("Error creating client")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	err = a.Storer.AddRedirectURIs(r.Context(), client.ID, redirectURIs)
	if err != nil {
		if e, ok := err.(clients.ErrRedirectURIAlreadyExists); ok {
			pos := -1
			for i, u := range redirectURIs {
				if u.URI == e.URI {
					pos = i
					break
				}
			}
			if pos < 0 {
				log := yall.FromContext(r.Context())
				log = log.WithField("err_uri", e.URI)
				log = log.WithField("passed_uris", redirectURIs)
				log.Error("source of ErrRedirectURIAlreadyExists wasn't a passed URI")
				api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
				return
			}
			api.Encode(w, r, http.StatusBadRequest, Response{Errors: []api.RequestError{{Field: "/redirectURIs/" + strconv.Itoa(pos) + "/id", Slug: api.RequestErrConflict}}})
			return
		}
		yall.FromContext(r.Context()).WithError(err).Error("Error creating client")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	yall.FromContext(r.Context()).WithField("client_id", client.ID).Debug("client created")
	api.Encode(w, r, http.StatusCreated, Response{Clients: []Client{apiClient(client)}, RedirectURIs: apiRedirectURIs(redirectURIs)})
}

func (a APIv1) handleGetClient(w http.ResponseWriter, r *http.Request) {
	input, resp := a.VerifyRequest(r)
	if resp != nil {
		api.Encode(w, r, resp.Status, resp)
		return
	}
	_ = input
	// TODO: retrieve client
}

func (a APIv1) handleDeleteClient(w http.ResponseWriter, r *http.Request) {
	input, resp := a.VerifyRequest(r)
	if resp != nil {
		api.Encode(w, r, resp.Status, resp)
		return
	}
	_ = input
	// TODO: delete client and redirect URIs
}

func (a APIv1) handleResetClientSecret(w http.ResponseWriter, r *http.Request) {
	input, resp := a.VerifyRequest(r)
	if resp != nil {
		api.Encode(w, r, resp.Status, resp)
		return
	}
	_ = input
	// TODO: reset client secret to new value
}

func (a APIv1) handleListClientRedirectURIs(w http.ResponseWriter, r *http.Request) {
	input, resp := a.VerifyRequest(r)
	if resp != nil {
		api.Encode(w, r, resp.Status, resp)
		return
	}
	_ = input
	// TODO: list client redirect URIs
}

func (a APIv1) handleCreateClientRedirectURIs(w http.ResponseWriter, r *http.Request) {
	input, resp := a.VerifyRequest(r)
	if resp != nil {
		api.Encode(w, r, resp.Status, resp)
		return
	}
	_ = input
	// TODO: add new client redirect URI
}

func (a APIv1) handleDeleteClientRedirectURI(w http.ResponseWriter, r *http.Request) {
	input, resp := a.VerifyRequest(r)
	if resp != nil {
		api.Encode(w, r, resp.Status, resp)
		return
	}
	_ = input
	// TODO: remove client redirect URI
}
