package reshape

import (
	"context"
	"encoding/base64"
	"time"

	openapi "github.com/ethanrous/weblens/api"
	auth_model "github.com/ethanrous/weblens/models/auth"
	"github.com/ethanrous/weblens/modules/structs"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TokenToTokenInfo(ctx context.Context, t *auth_model.Token) structs.TokenInfo {
	tokenStr := base64.StdEncoding.EncodeToString(t.Token[:])

	return structs.TokenInfo{
		Id:          t.Id.Hex(),
		CreatedTime: t.CreatedTime.UnixMilli(),
		LastUsed:    t.LastUsed.UnixMilli(),
		Nickname:    t.Nickname,
		Owner:       t.Owner,
		RemoteUsing: t.RemoteUsing,
		CreatedBy:   t.CreatedBy,
		Token:       tokenStr,
	}
}

func TokenInfoToToken(ctx context.Context, t openapi.TokenInfo) (*auth_model.Token, error) {
	id, _ := primitive.ObjectIDFromHex(t.Id)

	tokenSlice, err := base64.StdEncoding.DecodeString(t.Token)
	if err != nil {
		return nil, err
	}

	var token [32]byte

	copy(token[:], tokenSlice)

	return &auth_model.Token{
		Id:          id,
		CreatedTime: time.UnixMilli(t.CreatedTime),
		LastUsed:    time.UnixMilli(t.LastUsed),
		Nickname:    t.Nickname,
		Owner:       t.Owner,
		RemoteUsing: t.RemoteUsing,
		CreatedBy:   t.CreatedBy,
		Token:       token,
	}, nil
}

func TokensToTokenInfos(ctx context.Context, tokens []*auth_model.Token) []structs.TokenInfo {
	tokenInfos := make([]structs.TokenInfo, len(tokens))
	for i, t := range tokens {
		tokenInfos[i] = TokenToTokenInfo(ctx, t)
	}
	return tokenInfos
}
