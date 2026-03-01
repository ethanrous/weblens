// Package restuser provides REST API handlers for user operations.
package restuser

import (
	"net/http"

	auth_model "github.com/ethanrous/weblens/models/auth"
	"github.com/ethanrous/weblens/modules/netwrk"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/modules/wlstructs"
	"github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/reshape"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateAPIKey godoc
//
//	@ID			CreateAPIKey
//
//	@Security	SessionAuth
//
//	@Summary	Create a new api key
//	@Tags		APIKeys
//	@Produce	json
//
//	@Param		params	body		structs.APIKeyParams	true	"The new token params"
//
//	@Success	200		{object}	structs.TokenInfo		"The new token"
//	@Failure	403
//	@Failure	500
//	@Router		/keys [post]
func CreateAPIKey(ctx ctxservice.RequestContext) {
	tokenParams, err := netwrk.ReadRequestBody[wlstructs.APIKeyParams](ctx.Req)
	if err != nil {
		return
	}

	token, err := auth_model.GenerateNewToken(ctx, tokenParams.Name, ctx.Requester.Username, ctx.LocalTowerID)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.JSON(http.StatusOK, reshape.TokenToTokenInfo(ctx, token))
}

// GetMyTokens godoc
//
//	@ID	GetAPIKeys
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Get all api keys
//	@Tags		APIKeys
//	@Produce	json
//
//	@Success	200	{array}	structs.TokenInfo	"Tokens"
//	@Failure	403
//	@Failure	500
//	@Router		/keys [get]
func GetMyTokens(ctx ctxservice.RequestContext) {
	tokens, err := auth_model.GetTokensByUser(ctx, ctx.Requester.Username)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.JSON(http.StatusOK, reshape.TokensToTokenInfos(ctx, tokens))
}

// DeleteToken godoc
//
//	@ID	DeleteAPIKey
//
//	@Security
//	@Security	SessionAuth[admin]
//
//	@Summary	Delete an api key
//	@Tags		APIKeys
//	@Produce	json
//
//	@Param		tokenID	path	string	true	"API key id"
//
//	@Success	200
//	@Failure	403
//	@Failure	404
//	@Failure	500
//	@Router		/keys/{tokenID} [delete]
func DeleteToken(ctx ctxservice.RequestContext) {
	tokenID := ctx.Path("tokenID")

	objToken, err := primitive.ObjectIDFromHex(tokenID)
	if err != nil {
		ctx.Error(http.StatusBadRequest, err)

		return
	}

	// Check if the token exists
	_, err = auth_model.GetTokenByID(ctx, objToken)
	if err != nil {
		if wlerrors.Is(err, auth_model.ErrTokenNotFound) {
			ctx.Status(http.StatusNotFound)

			return
		}

		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	// Delete the token
	err = auth_model.DeleteToken(ctx, objToken)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.Status(http.StatusOK)
}
