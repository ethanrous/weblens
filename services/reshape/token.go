package reshape

import (
	"context"
	"encoding/base64"
	"time"

	auth_model "github.com/ethanrous/weblens/models/auth"
	"github.com/ethanrous/weblens/modules/structs"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TokenToTokenInfo converts an auth Token to a TokenInfo structure suitable for API responses.
func TokenToTokenInfo(_ context.Context, t *auth_model.Token) structs.TokenInfo {
	tokenStr := base64.StdEncoding.EncodeToString(t.Token[:])

	return structs.TokenInfo{
		ID:          t.ID.Hex(),
		CreatedTime: t.CreatedTime.UnixMilli(),
		LastUsed:    t.LastUsed.UnixMilli(),
		Nickname:    t.Nickname,
		Owner:       t.Owner,
		RemoteUsing: t.RemoteUsing,
		CreatedBy:   t.CreatedBy,
		Token:       tokenStr,
	}
}

// TokenInfoToToken converts a TokenInfo from the API to an auth Token.
func TokenInfoToToken(_ context.Context, t structs.TokenInfo) (*auth_model.Token, error) {
	id, _ := primitive.ObjectIDFromHex(t.ID)

	tokenSlice, err := base64.StdEncoding.DecodeString(t.Token)
	if err != nil {
		return nil, err
	}

	var token [32]byte

	copy(token[:], tokenSlice)

	return &auth_model.Token{
		ID:          id,
		CreatedTime: time.UnixMilli(t.CreatedTime),
		LastUsed:    time.UnixMilli(t.LastUsed),
		Nickname:    t.Nickname,
		Owner:       t.Owner,
		RemoteUsing: t.RemoteUsing,
		CreatedBy:   t.CreatedBy,
		Token:       token,
	}, nil
}

// TokensToTokenInfos converts a slice of auth Tokens to a slice of TokenInfo structures suitable for API responses.
func TokensToTokenInfos(ctx context.Context, tokens []*auth_model.Token) []structs.TokenInfo {
	tokenInfos := make([]structs.TokenInfo, len(tokens))
	for i, t := range tokens {
		tokenInfos[i] = TokenToTokenInfo(ctx, t)
	}

	return tokenInfos
}
