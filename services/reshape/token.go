package reshape

import (
	"context"

	auth_model "github.com/ethanrous/weblens/models/auth"
	"github.com/ethanrous/weblens/modules/structs"
)

func TokenToTokenInfo(ctx context.Context, t *auth_model.Token) *structs.Token {
	return &structs.Token{
		CreatedTime: t.CreatedTime,
		LastUsed:    t.LastUsed,
		Nickname:    t.Nickname,
		Owner:       t.Owner,
		RemoteUsing: t.RemoteUsing,
		CreatedBy:   t.CreatedBy,
		Token:       t.Token,
	}
}

func TokensToTokenInfos(ctx context.Context, tokens []*auth_model.Token) []*structs.Token {
	tokenInfos := make([]*structs.Token, len(tokens))
	for i, t := range tokens {
		tokenInfos[i] = TokenToTokenInfo(ctx, t)
	}
	return tokenInfos
}
