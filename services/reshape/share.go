package reshape

import (
	"context"

	share_model "github.com/ethanrous/weblens/models/share"
	user_model "github.com/ethanrous/weblens/models/usermodel"
	"github.com/ethanrous/weblens/modules/wlog"
	"github.com/ethanrous/weblens/modules/wlstructs"
)

// ShareToShareInfo converts a FileShare model to a ShareInfo transfer object.
func ShareToShareInfo(ctx context.Context, s *share_model.FileShare) wlstructs.ShareInfo {
	accessors := make([]wlstructs.UserInfo, 0, len(s.Accessors))

	for _, a := range s.Accessors {
		u, err := user_model.GetUserByUsername(ctx, a)
		if err != nil {
			wlog.FromContext(ctx).Error().Stack().Err(err).Str("username", a).Msg("failed to get user by username")

			continue
		}

		accessors = append(accessors, UserToUserInfo(ctx, u))
	}

	id := s.ShareID.Hex()
	if s.ShareID.IsZero() {
		id = ""
	}

	return wlstructs.ShareInfo{
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
func PermissionsToPermissionsInfo(_ context.Context, perms map[string]*share_model.Permissions) map[string]wlstructs.PermissionsInfo {
	permsInfo := make(map[string]wlstructs.PermissionsInfo, len(perms))
	for k, v := range perms {
		permsInfo[k] = toPermissionInfo(*v)
	}

	return permsInfo
}

func toPermissionInfo(perms share_model.Permissions) wlstructs.PermissionsInfo {
	return wlstructs.PermissionsInfo{
		CanView:     perms.CanView,
		CanEdit:     perms.CanEdit,
		CanDownload: perms.CanDownload,
		CanDelete:   perms.CanDelete,
	}
}

// PermissionsParamsToPermissions converts PermissionsParams to a Permissions model.
func PermissionsParamsToPermissions(_ context.Context, perms wlstructs.PermissionsParams) (share_model.Permissions, error) {
	newPerms := share_model.Permissions{
		CanView:     perms.CanView,
		CanEdit:     perms.CanEdit,
		CanDownload: perms.CanDownload,
		CanDelete:   perms.CanDelete,
	}

	// Ensure that CanView is always true, for now.
	// This may be revisited in the future for more granular control.
	newPerms.CanView = true

	return newPerms, nil
}

// UnpackNewUserParams extracts the username and permissions from AddUserParams.
func UnpackNewUserParams(ctx context.Context, params wlstructs.AddUserParams) (string, share_model.Permissions, error) {
	perms, err := PermissionsParamsToPermissions(ctx, params.PermissionsParams)
	if err != nil {
		return "", share_model.Permissions{}, err
	}

	return params.Username, perms, nil
}
