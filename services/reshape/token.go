package reshape

import (
	"context"
	"time"

	auth_model "github.com/ethanrous/weblens/models/auth"
	"github.com/ethanrous/weblens/modules/structs"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TokenToTokenInfo(ctx context.Context, t *auth_model.Token) structs.TokenInfo {
	return structs.TokenInfo{
		Id:          t.Id.Hex(),
		CreatedTime: t.CreatedTime.UnixMilli(),
		LastUsed:    t.LastUsed.UnixMilli(),
		Nickname:    t.Nickname,
		Owner:       t.Owner,
		RemoteUsing: t.RemoteUsing,
		CreatedBy:   t.CreatedBy,
		Token:       t.Token,
	}
}

func TokenInfoToToken(ctx context.Context, t structs.TokenInfo) *auth_model.Token {
	id, _ := primitive.ObjectIDFromHex(t.Id)

	return &auth_model.Token{
		Id:          id,
		CreatedTime: time.UnixMilli(t.CreatedTime),
		LastUsed:    time.UnixMilli(t.LastUsed),
		Nickname:    t.Nickname,
		Owner:       t.Owner,
		RemoteUsing: t.RemoteUsing,
		CreatedBy:   t.CreatedBy,
		Token:       t.Token,
	}
}

func TokensToTokenInfos(ctx context.Context, tokens []*auth_model.Token) []structs.TokenInfo {
	tokenInfos := make([]structs.TokenInfo, len(tokens))
	for i, t := range tokens {
		tokenInfos[i] = TokenToTokenInfo(ctx, t)
	}
	return tokenInfos
}
