package reshape

import (
	"context"

	share_model "github.com/ethanrous/weblens/models/share"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/modules/structs"
)

// ShareToShareInfo converts a FileShare model to a ShareInfo transfer object.
func ShareToShareInfo(ctx context.Context, s *share_model.FileShare) structs.ShareInfo {
	accessors := make([]structs.UserInfo, 0, len(s.Accessors))

	for _, a := range s.Accessors {
		u, err := user_model.GetUserByUsername(ctx, a)
		if err != nil {
			log.FromContext(ctx).Error().Stack().Err(err).Str("username", a).Msg("failed to get user by username")

			continue
		}

		accessors = append(accessors, UserToUserInfo(ctx, u))
	}

	id := s.ShareID.Hex()
	if s.ShareID.IsZero() {
		id = ""
	}

	return structs.ShareInfo{
		ShareID:     id,
		FileID:      s.FileID,
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

// PermissionsToPermissionsInfo converts a map of Permissions models to PermissionsInfo transfer objects.
func PermissionsToPermissionsInfo(_ context.Context, perms map[string]*share_model.Permissions) map[string]structs.PermissionsInfo {
	permsInfo := make(map[string]structs.PermissionsInfo, len(perms))
	for k, v := range perms {
		permsInfo[k] = structs.PermissionsInfo{
			CanView:     v.CanView,
			CanEdit:     v.CanEdit,
			CanDownload: v.CanDownload,
			CanDelete:   v.CanDelete,
		}
	}

	return permsInfo
}

// PermissionsParamsToPermissions converts PermissionsParams to a Permissions model.
func PermissionsParamsToPermissions(_ context.Context, perms structs.PermissionsParams) (share_model.Permissions, error) {
	newPerms := share_model.Permissions{
		CanView:     perms.CanView,
		CanEdit:     perms.CanEdit,
		CanDownload: perms.CanDownload,
		CanDelete:   perms.CanDelete,
	}

	return newPerms, nil
}

// UnpackNewUserParams extracts the username and permissions from AddUserParams.
func UnpackNewUserParams(ctx context.Context, params structs.AddUserParams) (string, share_model.Permissions, error) {
	perms, err := PermissionsParamsToPermissions(ctx, params.PermissionsParams)
	if err != nil {
		return "", share_model.Permissions{}, err
	}

	return params.Username, perms, nil
}
