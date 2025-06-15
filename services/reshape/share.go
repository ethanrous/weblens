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
			ctx.Log().Error().Stack().Err(err).Str("username", a).Msg("failed to get user by username")

			continue
		}

		accessors = append(accessors, UserToUserInfo(ctx, u))
	}

	id := s.ShareId.Hex()
	if s.ShareId.IsZero() {
		id = ""
	}

	return structs.ShareInfo{
		ShareId:     id,
		FileId:      s.FileId,
		ShareName:   s.ShareName,
		Owner:       s.Owner,
		Accessors:   accessors,
		Permissions: PermissionsToPermissionsInfo(ctx, s.Permissions),
		Public:      s.Public,
		Wormhole:    s.Wormhole,
		Enabled:     s.Enabled,
		Expires:     s.Expires.UnixMilli(),
		Updated:     s.Updated.UnixMilli(),
	}
}

func PermissionsToPermissionsInfo(ctx context.RequestContext, perms map[string]*share_model.Permissions) map[string]structs.PermissionsInfo {
	permsInfo := make(map[string]structs.PermissionsInfo, len(perms))
	for k, v := range perms {
		permsInfo[k] = structs.PermissionsInfo{
			CanEdit:     v.CanEdit,
			CanDownload: v.CanDownload,
			CanDelete:   v.CanDelete,
		}
	}

	return permsInfo
}

func PermissionsParamsToPermissions(ctx context.RequestContext, perms structs.PermissionsParams) (share_model.Permissions, error) {
	newPerms := share_model.Permissions{
		CanEdit:     perms.CanEdit,
		CanDownload: perms.CanDownload,
		CanDelete:   perms.CanDelete,
	}

	return newPerms, nil
}

func UnpackNewUserParams(ctx context.RequestContext, params structs.AddUserParams) (string, share_model.Permissions, error) {
	perms, err := PermissionsParamsToPermissions(ctx, params.PermissionsParams)
	if err != nil {
		return "", share_model.Permissions{}, err
	}

	return params.Username, perms, nil
}
