package apiv1

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"darlinggo.co/api"
	uuid "github.com/hashicorp/go-uuid"
	"impractical.co/auth/clients"
	"impractical.co/userip"
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
	for pos, uri := range body.RedirectURIs {
		if uri.URI == "" {
			reqErrs = append(reqErrs, api.RequestError{Field: fmt.Sprintf("/redirectURIs/%d/URI", pos), Slug: api.RequestErrMissing})
		}
	}
	if len(reqErrs) > 0 {
		api.Encode(w, r, http.StatusBadRequest, reqErrs)
		return
	}
	id, err := uuid.GenerateUUID()
	if err != nil {
		yall.FromContext(r.Context()).WithError(err).Error("Error creating client ID")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	body.Client.ID = id
	body.Client.CreatedAt = time.Now()
	body.Client.CreatedBy = a.Signer.Key
	body.Client.CreatedByIP = userip.Get(r)
	if body.Client.CreatedByIP == "" {
		yall.FromContext(r.Context()).Error("Couldn't determine user's IP")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	if body.Client.Confidential {
		b := make([]byte, 16)
		_, err := rand.Read(b)
		if err != nil {
			yall.FromContext(r.Context()).Error("Couldn't generate client secret")
			api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
			return
		}
		body.Client.Secret = hex.EncodeToString(b)
	}
	for pos, uri := range body.RedirectURIs {
		id, err := uuid.GenerateUUID()
		if err != nil {
			yall.FromContext(r.Context()).WithError(err).Error("Error creating redirect URI ID")
			api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
			return
		}
		uri.ID = id
		uri.ClientID = body.Client.ID
		uri.CreatedAt = body.Client.CreatedAt
		uri.CreatedBy = body.Client.CreatedBy
		uri.CreatedByIP = body.Client.CreatedByIP
		body.RedirectURIs[pos] = uri
	}
	client := coreClient(body.Client)
	ch, err := clients.ChangeSecret([]byte(body.Client.Secret))
	if err != nil {
		yall.FromContext(r.Context()).WithError(err).Error("Error setting client secret")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	clients.Apply(ch, client)
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
	err = a.Storer.AddRedirectURIs(r.Context(), redirectURIs)
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
	respClient := apiClient(client)
	respClient.Secret = body.Client.Secret
	api.Encode(w, r, http.StatusCreated, Response{Clients: []Client{respClient}, RedirectURIs: apiRedirectURIs(redirectURIs)})
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
