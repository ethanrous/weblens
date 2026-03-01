package reshape

import (
	"context"

	"github.com/ethanrous/weblens/models/usermodel"
	user_model "github.com/ethanrous/weblens/models/usermodel"
	"github.com/ethanrous/weblens/modules/wlstructs"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
)

// UserToUserInfo converts a User model to a UserInfo transfer object.
func UserToUserInfo(ctx context.Context, u *user_model.User) wlstructs.UserInfo {
	userIsOnline := getUserIsOnline(ctx, u.Username)

	return wlstructs.UserInfo{
		Username:        u.Username,
		FullName:        u.DisplayName,
		HomeID:          u.HomeID,
		TrashID:         u.TrashID,
		PermissionLevel: int(u.UserPerms),
		Activated:       u.Activated,
		IsOnline:        userIsOnline,
		UpdatedAt:       u.UpdatedAt,
	}
}

// UserToUserInfoArchive converts a User model to a UserInfoArchive transfer object for backup purposes.
func UserToUserInfoArchive(ctx context.Context, u *usermodel.User) wlstructs.UserInfoArchive {
	if u == nil || u.IsSystemUser() {
		return wlstructs.UserInfoArchive{}
	}

	userIsOnline := getUserIsOnline(ctx, u.Username)

	info := wlstructs.UserInfoArchive{
		UserInfo: wlstructs.UserInfo{
			Username:        u.GetUsername(),
			FullName:        u.DisplayName,
			PermissionLevel: int(u.UserPerms),
			HomeID:          u.HomeID,
			TrashID:         u.TrashID,
			Activated:       u.IsActive(),
			IsOnline:        userIsOnline,
			UpdatedAt:       u.UpdatedAt,
		},
		Password: u.Password,
	}

	return info
}

// UserInfoArchiveToUser converts a UserInfoArchive transfer object to a User model for restoration.
func UserInfoArchiveToUser(uInfo wlstructs.UserInfoArchive) *usermodel.User {
	u := &usermodel.User{
		Username:  uInfo.Username,
		Password:  uInfo.Password,
		Activated: uInfo.Activated,
		UserPerms: user_model.Permissions(uInfo.PermissionLevel),
		HomeID:    uInfo.HomeID,
		TrashID:   uInfo.TrashID,
		UpdatedAt: uInfo.UpdatedAt,
	}

	return u
}

func getUserIsOnline(ctx context.Context, username string) bool {
	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return false
	}

	client := appCtx.ClientService.GetClientByUsername(username)

	return client != nil
}
