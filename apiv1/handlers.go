package apiv1

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"darlinggo.co/api"
	"darlinggo.co/trout/v2"
	uuid "github.com/hashicorp/go-uuid"
	"impractical.co/userip"
	yall "yall.in"

	"lockbox.dev/clients"
)

func (a APIv1) handleCreateClient(w http.ResponseWriter, r *http.Request) {
	input, resp := a.VerifyRequest(r)
	if resp != nil {
		api.Encode(w, r, resp.Status, resp)
		return
	}
	var body Client
	err := json.Unmarshal([]byte(input), &body)
	if err != nil {
		yall.FromContext(r.Context()).WithError(err).Debug("Error decoding request body")
		api.Encode(w, r, http.StatusBadRequest, Response{Errors: api.InvalidFormatError})
		return
	}

	id, err := uuid.GenerateUUID()
	if err != nil {
		yall.FromContext(r.Context()).WithError(err).Error("Error creating client ID")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	body.ID = id
	body.CreatedAt = time.Now()
	body.CreatedBy = a.Signer.Key
	body.CreatedByIP = userip.Get(r)
	if body.CreatedByIP == "" {
		yall.FromContext(r.Context()).Error("Couldn't determine user's IP")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	if body.Confidential {
		secretBytes := make([]byte, 16) //nolint:gomnd // chosen arbitrarily, doesn't matter if it changes
		_, err = rand.Read(secretBytes)
		if err != nil {
			yall.FromContext(r.Context()).WithError(err).Error("Couldn't generate client secret")
			api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
			return
		}
		body.Secret = hex.EncodeToString(secretBytes)
	}
	client := coreClient(body)
	change, err := clients.ChangeSecret([]byte(body.Secret))
	if err != nil {
		yall.FromContext(r.Context()).WithError(err).Error("Error setting client secret")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	client = clients.Apply(change, client)
	err = a.Storer.Create(r.Context(), client)
	if err != nil {
		if errors.Is(err, clients.ErrClientAlreadyExists) {
			api.Encode(w, r, http.StatusBadRequest, Response{Errors: []api.RequestError{{Field: "/client/id", Slug: api.RequestErrConflict}}})
			return
		}
		yall.FromContext(r.Context()).WithError(err).Error("Error creating client")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	yall.FromContext(r.Context()).WithField("client_id", client.ID).Debug("client created")
	respClient := apiClient(client)
	respClient.Secret = body.Secret
	api.Encode(w, r, http.StatusCreated, Response{Clients: []Client{respClient}})
}

func (a APIv1) handleGetClient(w http.ResponseWriter, r *http.Request) {
	if _, resp := a.VerifyRequest(r); resp != nil {
		api.Encode(w, r, resp.Status, resp)
		return
	}
	vars := trout.RequestVars(r)
	id := vars.Get("id")
	if id == "" {
		api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Param: "id", Slug: api.RequestErrMissing}}})
		return
	}
	client, err := a.Storer.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, clients.ErrClientNotFound) {
			api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Param: "id", Slug: api.RequestErrNotFound}}})
			return
		}
		yall.FromContext(r.Context()).WithError(err).Error("error retrieving client")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	yall.FromContext(r.Context()).WithField("client_id", client.ID).Debug("Client retrieved")
	api.Encode(w, r, http.StatusOK, Response{Clients: []Client{apiClient(client)}})
}

func (a APIv1) handleDeleteClient(w http.ResponseWriter, r *http.Request) {
	if _, resp := a.VerifyRequest(r); resp != nil {
		api.Encode(w, r, resp.Status, resp)
		return
	}
	vars := trout.RequestVars(r)
	clientID := vars.Get("id")
	if clientID == "" {
		api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Param: "id", Slug: api.RequestErrMissing}}})
		return
	}
	client, err := a.Storer.Get(r.Context(), clientID)
	if err != nil {
		if errors.Is(err, clients.ErrClientNotFound) {
			api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Param: "id", Slug: api.RequestErrNotFound}}})
			return
		}
		yall.FromContext(r.Context()).WithError(err).Error("error retrieving client")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	redirectURIs, err := a.Storer.ListRedirectURIs(r.Context(), clientID)
	if err != nil {
		yall.FromContext(r.Context()).WithError(err).Error("error listing redirect URIs")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	if len(redirectURIs) > 0 {
		ids := make([]string, 0, len(redirectURIs))
		for _, uri := range redirectURIs {
			ids = append(ids, uri.ID)
		}
		err = a.Storer.RemoveRedirectURIs(r.Context(), ids)
		if err != nil {
			yall.FromContext(r.Context()).WithError(err).Error("error removing reidrect URIs")
			api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
			return
		}
	}
	err = a.Storer.Delete(r.Context(), clientID)
	if err != nil {
		if errors.Is(err, clients.ErrClientNotFound) {
			api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Param: "id", Slug: api.RequestErrNotFound}}})
			return
		}
		yall.FromContext(r.Context()).WithError(err).Error("error deleting client")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	yall.FromContext(r.Context()).WithField("client_id", client.ID).Debug("Client deleted")
	api.Encode(w, r, http.StatusOK, Response{Clients: []Client{apiClient(client)}, RedirectURIs: apiRedirectURIs(redirectURIs)})
}

func (a APIv1) handleResetClientSecret(w http.ResponseWriter, r *http.Request) {
	if _, resp := a.VerifyRequest(r); resp != nil {
		api.Encode(w, r, resp.Status, resp)
		return
	}
	vars := trout.RequestVars(r)
	clientID := vars.Get("id")
	if clientID == "" {
		api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Param: "id", Slug: api.RequestErrMissing}}})
		return
	}
	client, err := a.Storer.Get(r.Context(), clientID)
	if err != nil {
		if errors.Is(err, clients.ErrClientNotFound) {
			api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Param: "id", Slug: api.RequestErrNotFound}}})
			return
		}
		yall.FromContext(r.Context()).WithError(err).Error("error retrieving client")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	secretBytes := make([]byte, 16) //nolint:gomnd // chosen arbitrarily, doesn't matter if it changes
	_, err = rand.Read(secretBytes)
	if err != nil {
		yall.FromContext(r.Context()).WithError(err).Error("Couldn't generate client secret")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	respClient := apiClient(client)
	respClient.Secret = hex.EncodeToString(secretBytes)
	change, err := clients.ChangeSecret([]byte(respClient.Secret))
	if err != nil {
		yall.FromContext(r.Context()).WithError(err).Error("Error setting client secret")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	err = a.Storer.Update(r.Context(), clientID, change)
	if err != nil {
		if errors.Is(err, clients.ErrClientNotFound) {
			api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Param: "id", Slug: api.RequestErrNotFound}}})
			return
		}
		yall.FromContext(r.Context()).WithError(err).Error("Error updating client secret")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	yall.FromContext(r.Context()).WithField("client_id", client.ID).Debug("updated client secret")
	api.Encode(w, r, http.StatusOK, Response{Clients: []Client{respClient}})
}

func (a APIv1) handleListClientRedirectURIs(w http.ResponseWriter, r *http.Request) {
	if _, resp := a.VerifyRequest(r); resp != nil {
		api.Encode(w, r, resp.Status, resp)
		return
	}
	vars := trout.RequestVars(r)
	clientID := vars.Get("id")
	if clientID == "" {
		api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Param: "id", Slug: api.RequestErrMissing}}})
		return
	}
	_, err := a.Storer.Get(r.Context(), clientID)
	if err != nil {
		if errors.Is(err, clients.ErrClientNotFound) {
			api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Param: "id", Slug: api.RequestErrNotFound}}})
			return
		}
		yall.FromContext(r.Context()).WithError(err).Error("error retrieving client")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	redirectURIs, err := a.Storer.ListRedirectURIs(r.Context(), clientID)
	if err != nil {
		yall.FromContext(r.Context()).WithError(err).Error("error listing redirect URIs")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	yall.FromContext(r.Context()).WithField("client_id", clientID).Debug("redirect URIs retrieved")
	api.Encode(w, r, http.StatusOK, Response{RedirectURIs: apiRedirectURIs(redirectURIs)})
}

func (a APIv1) handleCreateClientRedirectURIs(w http.ResponseWriter, r *http.Request) {
	input, resp := a.VerifyRequest(r)
	if resp != nil {
		api.Encode(w, r, resp.Status, resp)
		return
	}
	vars := trout.RequestVars(r)
	clientID := vars.Get("id")
	if clientID == "" {
		api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Param: "id", Slug: api.RequestErrMissing}}})
		return
	}
	_, err := a.Storer.Get(r.Context(), clientID)
	if err != nil {
		if errors.Is(err, clients.ErrClientNotFound) {
			api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Param: "id", Slug: api.RequestErrNotFound}}})
			return
		}
		yall.FromContext(r.Context()).WithError(err).Error("error retrieving client")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}

	var body struct {
		RedirectURIs []RedirectURI `json:"redirectURIs"`
	}
	err = json.Unmarshal([]byte(input), &body)
	if err != nil {
		yall.FromContext(r.Context()).WithError(err).Debug("Error decoding request body")
		api.Encode(w, r, http.StatusBadRequest, Response{Errors: api.InvalidFormatError})
		return
	}

	var reqErrs []api.RequestError
	for pos, uri := range body.RedirectURIs {
		if uri.URI == "" {
			reqErrs = append(reqErrs, api.RequestError{Field: fmt.Sprintf("/redirectURIs/%d/URI", pos), Slug: api.RequestErrMissing})
		}
	}
	if len(reqErrs) > 0 {
		api.Encode(w, r, http.StatusBadRequest, reqErrs)
		return
	}
	createdAt := time.Now()
	createdByIP := userip.Get(r)
	if createdByIP == "" {
		yall.FromContext(r.Context()).Error("Couldn't determine user's IP")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	for pos, uri := range body.RedirectURIs {
		var id string
		id, err = uuid.GenerateUUID()
		if err != nil {
			yall.FromContext(r.Context()).WithError(err).Error("Error creating redirect URI ID")
			api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
			return
		}
		uri.ID = id
		uri.ClientID = clientID
		uri.CreatedAt = createdAt
		uri.CreatedBy = a.Signer.Key
		uri.CreatedByIP = createdByIP
		body.RedirectURIs[pos] = uri
	}
	redirectURIs := coreRedirectURIs(body.RedirectURIs)
	err = a.Storer.AddRedirectURIs(r.Context(), redirectURIs)
	if err != nil {
		var uriAlreadyExistsErr clients.RedirectURIAlreadyExistsError
		if errors.As(err, &uriAlreadyExistsErr) {
			pos := findRedirectURIErrorPos(r.Context(), redirectURIs, uriAlreadyExistsErr)
			if pos < 0 {
				log := yall.FromContext(r.Context())
				log = log.WithField("err_uri", uriAlreadyExistsErr.URI)
				log = log.WithField("passed_uris", redirectURIs)
				log.Error("source of ErrRedirectURIAlreadyExists wasn't a passed URI")
				api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
				return
			}
			api.Encode(w, r, http.StatusBadRequest, Response{Errors: []api.RequestError{{Field: "/redirectURIs/" + strconv.Itoa(pos) + "/id", Slug: api.RequestErrConflict}}})
			return
		}
		yall.FromContext(r.Context()).WithError(err).Error("Error creating redirect URIs")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	yall.FromContext(r.Context()).WithField("client_id", clientID).Debug("redirect URIs added")
	api.Encode(w, r, http.StatusCreated, Response{RedirectURIs: apiRedirectURIs(redirectURIs)})
}

func (a APIv1) handleDeleteClientRedirectURI(w http.ResponseWriter, r *http.Request) {
	if _, resp := a.VerifyRequest(r); resp != nil {
		api.Encode(w, r, resp.Status, resp)
		return
	}
	vars := trout.RequestVars(r)
	clientID := vars.Get("id")
	if clientID == "" {
		api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Param: "id", Slug: api.RequestErrMissing}}})
		return
	}
	uriID := vars.Get("uri")
	if uriID == "" {
		api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Param: "uri", Slug: api.RequestErrMissing}}})
		return
	}
	_, err := a.Storer.Get(r.Context(), clientID)
	if err != nil {
		if errors.Is(err, clients.ErrClientNotFound) {
			api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Param: "id", Slug: api.RequestErrNotFound}}})
			return
		}
		yall.FromContext(r.Context()).WithError(err).Error("error retrieving client")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	redirectURIs, err := a.Storer.ListRedirectURIs(r.Context(), clientID)
	if err != nil {
		yall.FromContext(r.Context()).WithError(err).Error("error listing redirect URIs")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	var redirectURI RedirectURI
	for _, uri := range redirectURIs {
		if uri.ID == uriID {
			redirectURI = apiRedirectURI(uri)
			break
		}
	}
	if redirectURI.ID == "" {
		yall.FromContext(r.Context()).WithField("client_id", clientID).WithField("redirect_uri_id", uriID).Debug("redirect URI not found in client")
		api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Param: "uri", Slug: api.RequestErrNotFound}}})
		return
	}
	err = a.Storer.RemoveRedirectURIs(r.Context(), []string{redirectURI.ID})
	if err != nil {
		yall.FromContext(r.Context()).WithField("client_id", clientID).WithField("redirect_uri_id", uriID).WithError(err).Error("error removing redirect URI")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	yall.FromContext(r.Context()).WithField("client_id", clientID).WithField("redirect_uri_id", uriID).Debug("redirect URI removed")
	api.Encode(w, r, http.StatusOK, Response{RedirectURIs: []RedirectURI{redirectURI}})
}

func findRedirectURIErrorPos(_ context.Context, uris []clients.RedirectURI, uriErr clients.RedirectURIAlreadyExistsError) int {
	for i, u := range uris {
		if u.URI == uriErr.URI {
			return i
		}
	}
	return -1
}
