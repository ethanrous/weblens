package user

import (
	"net/http"

	auth_model "github.com/ethanrous/weblens/models/auth"
	"github.com/ethanrous/weblens/modules/net"
	"github.com/ethanrous/weblens/modules/structs"
	"github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/reshape"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
//	@Param		params	body		rest.ApiKeyParams	true	"The new token params"
//
//	@Success	200		{object}	structs.Token		"The new token"
//	@Failure	403
//	@Failure	500
//	@Router		/keys [post]
func CreateApiKey(ctx context.RequestContext) {

	tokenParams, err := net.ReadRequestBody[structs.ApiKeyParams](ctx.Req)
	if err != nil {
		return
	}

	token, err := auth_model.GenerateNewToken(ctx, tokenParams.Name, ctx.Requester.Username, ctx.LocalTowerId)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, reshape.TokenToTokenInfo(ctx, token))
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
//	@Success	200	{array}	structs.Token	"Tokens"
//	@Failure	403
//	@Failure	500
//	@Router		/keys [get]
func GetMyTokens(ctx context.RequestContext) {
	tokens, err := auth_model.GetTokensByUser(ctx, ctx.Requester.Username)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, reshape.TokensToTokenInfos(ctx, tokens))
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
//	@Router		/keys/{tokenId} [delete]
func DeleteToken(ctx context.RequestContext) {
	tokenId := ctx.Path("tokenId")

	objToken, err := primitive.ObjectIDFromHex(tokenId)
	if err != nil {
		ctx.Error(http.StatusBadRequest, err)
		return
	}

	// Check if the token exists
	_, err = auth_model.GetTokenById(ctx, objToken)
	if err != nil {
		if errors.Is(err, auth_model.ErrTokenNotFound) {
			ctx.W.WriteHeader(http.StatusNotFound)
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

	ctx.W.WriteHeader(http.StatusOK)
}
