package reshape

import (
	"context"

	auth_model "github.com/ethanrous/weblens/models/auth"
	"github.com/ethanrous/weblens/modules/structs"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TokenToTokenInfo(ctx context.Context, t *auth_model.Token) structs.Token {
	return structs.Token{
		Id:          t.Id.Hex(),
		CreatedTime: t.CreatedTime,
		LastUsed:    t.LastUsed,
		Nickname:    t.Nickname,
		Owner:       t.Owner,
		RemoteUsing: t.RemoteUsing,
		CreatedBy:   t.CreatedBy,
		Token:       t.Token,
	}
}

func TokenInfoToToken(ctx context.Context, t structs.Token) *auth_model.Token {
	id, _ := primitive.ObjectIDFromHex(t.Id)

	return &auth_model.Token{
		Id:          id,
		CreatedTime: t.CreatedTime,
		LastUsed:    t.LastUsed,
		Nickname:    t.Nickname,
		Owner:       t.Owner,
		RemoteUsing: t.RemoteUsing,
		CreatedBy:   t.CreatedBy,
		Token:       t.Token,
	}
}

func TokensToTokenInfos(ctx context.Context, tokens []*auth_model.Token) []structs.Token {
	tokenInfos := make([]structs.Token, len(tokens))
	for i, t := range tokens {
		tokenInfos[i] = TokenToTokenInfo(ctx, t)
	}
	return tokenInfos
}
