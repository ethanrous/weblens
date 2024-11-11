package http

import (
	"net/http"

	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/rest"
	"github.com/gin-gonic/gin"
)

// CreateApiKey godoc
//
//	@ID	CreateApiKey
//
//	@Security
//	@Security	SessionAuth[admin]
//
//	@Summary	Create a new api key
//	@Tags		ApiKeys
//	@Produce	json
//
//	@Success	200	{object}	rest.ApiKeyInfo	"The new api key info"
//	@Failure	403
//	@Failure	500
//	@Router		/keys [post]
func newApiKey(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)

	if !u.IsAdmin() {
		ctx.Status(http.StatusForbidden)
		return
	}

	newKey, err := pack.AccessService.GenerateApiKey(u, pack.InstanceService.GetLocal())
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	apiKeyInfo := rest.ApiKeyToApiKeyInfo(newKey)
	ctx.JSON(http.StatusOK, apiKeyInfo)
}

// GetApiKeys godoc
//
//	@ID	GetApiKeys
//
//	@Security
//	@Security	SessionAuth[admin]
//
//	@Summary	Get all api keys
//	@Tags		ApiKeys
//	@Produce	json
//
//	@Success	200	{array}	rest.ApiKeyInfo	"Api keys info"
//	@Failure	403
//	@Failure	500
//	@Router		/keys [get]
func getApiKeys(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)

	keys, err := pack.AccessService.GetAllKeys(u)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	var keysInfo []rest.ApiKeyInfo
	for _, key := range keys {
		keysInfo = append(keysInfo, rest.ApiKeyToApiKeyInfo(key))
	}

	ctx.JSON(http.StatusOK, keysInfo)
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
func deleteApiKey(ctx *gin.Context) {
	pack := getServices(ctx)
	key := models.WeblensApiKey(ctx.Param("keyId"))
	keyInfo, err := pack.AccessService.GetApiKey(key)
	if err != nil || keyInfo.Key == "" {
		log.ShowErr(err)
		ctx.Status(http.StatusNotFound)
		return
	}

	err = pack.AccessService.DeleteApiKey(key)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}
