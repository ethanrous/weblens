package reshape

import (
	share_model "github.com/ethanrous/weblens/models/share"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/structs"
	"github.com/ethanrous/weblens/services/context"
)

func ShareToShareInfo(ctx context.RequestContext, s *share_model.FileShare) structs.ShareInfo {
	accessors := make([]structs.UserInfo, 0, len(s.Accessors))
	for _, a := range s.Accessors {
		u, err := user_model.GetUserByUsername(ctx, a)
		if err != nil {
			ctx.Log.Error().Stack().Err(err).Str("username", a).Msg("failed to get user by username")
			continue
		}
		accessors = append(accessors, UserToUserInfo(ctx, u))
	}

	return structs.ShareInfo{
		ShareId:   s.ShareId,
		FileId:    s.FileId,
		ShareName: s.ShareName,
		Owner:     s.Owner,
		Accessors: accessors,
		Public:    s.Public,
		Wormhole:  s.Wormhole,
		Enabled:   s.Enabled,
		Expires:   s.Expires.UnixMilli(),
		Updated:   s.Updated.UnixMilli(),
	}
}
