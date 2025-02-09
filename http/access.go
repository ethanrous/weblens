package http

import (
	"net/http"

	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/rest"
	"github.com/go-chi/chi/v5"
)

// CreateApiKey godoc
//
//	@ID			CreateApiKey
//
//	@Security	SessionAuth
//
//	@Summary	Create a new api key
//	@Tags		ApiKeys
//	@Produce	json
//
//	@Param		params	body		rest.ApiKeyParams	true	"The new key params"
//
//	@Success	200		{object}	rest.ApiKeyInfo		"The new api key info"
//	@Failure	403
//	@Failure	500
//	@Router		/keys [post]
func newApiKey(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r, false)
	if SafeErrorAndExit(err, w) {
		return
	}

	body, err := readCtxBody[rest.ApiKeyParams](w, r)
	if err != nil {
		return
	}

	newKey, err := pack.AccessService.GenerateApiKey(u, pack.InstanceService.GetLocal(), body.Name)
	if err != nil {
		log.ShowErr(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	apiKeyInfo := rest.ApiKeyToApiKeyInfo(newKey)
	writeJson(w, http.StatusOK, apiKeyInfo)
}

// GetApiKeys godoc
//
//	@ID	GetApiKeys
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Get all api keys
//	@Tags		ApiKeys
//	@Produce	json
//
//	@Success	200	{array}	rest.ApiKeyInfo	"Api keys info"
//	@Failure	403
//	@Failure	500
//	@Router		/keys [get]
func getApiKeys(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r, false)
	if SafeErrorAndExit(err, w) {
		return
	}

	keys, err := pack.AccessService.GetKeysByUser(u)
	if SafeErrorAndExit(err, w) {
		return
	}

	var keysInfo []rest.ApiKeyInfo
	for _, key := range keys {
		keysInfo = append(keysInfo, rest.ApiKeyToApiKeyInfo(key))
	}

	writeJson(w, http.StatusOK, keysInfo)
}

// DeleteApiKey godoc
//
//	@ID	DeleteApiKey
//
//	@Security
//	@Security	SessionAuth[admin]
//
//	@Summary	Delete an api key
//	@Tags		ApiKeys
//	@Produce	json
//
//	@Param		keyId	path	string	true	"Api key id"
//
//	@Success	200
//	@Failure	403
//	@Failure	404
//	@Failure	500
//	@Router		/keys/{keyId} [delete]
func deleteApiKey(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	key := models.WeblensApiKey(chi.URLParam(r, "keyId"))
	keyInfo, err := pack.AccessService.GetApiKey(key)
	if err != nil || keyInfo.Key == "" {
		log.ShowErr(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = pack.AccessService.DeleteApiKey(key)
	if err != nil {
		log.ShowErr(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
