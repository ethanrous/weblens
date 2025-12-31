package reshape

import (
	"context"

	"github.com/ethanrous/weblens/models/user"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/structs"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
)

// UserToUserInfo converts a User model to a UserInfo transfer object.
func UserToUserInfo(ctx context.Context, u *user_model.User) structs.UserInfo {
	userIsOnline := getUserIsOnline(ctx, u.Username)

	return structs.UserInfo{
		Username:        u.Username,
		FullName:        u.DisplayName,
		HomeID:          u.HomeID,
		TrashID:         u.TrashID,
		PermissionLevel: int(u.UserPerms),
		Activated:       u.Activated,
		IsOnline:        userIsOnline,
	}
}

// UserToUserInfoArchive converts a User model to a UserInfoArchive transfer object for backup purposes.
func UserToUserInfoArchive(ctx context.Context, u *user.User) structs.UserInfoArchive {
	if u == nil || u.IsSystemUser() {
		return structs.UserInfoArchive{}
	}

	userIsOnline := getUserIsOnline(ctx, u.Username)

	info := structs.UserInfoArchive{
		UserInfo: structs.UserInfo{
			Username:        u.GetUsername(),
			FullName:        u.DisplayName,
			PermissionLevel: int(u.UserPerms),
			HomeID:          u.HomeID,
			TrashID:         u.TrashID,
			Activated:       u.IsActive(),
			IsOnline:        userIsOnline,
		},
		Password: u.Password,
	}

	return info
}

// UserInfoArchiveToUser converts a UserInfoArchive transfer object to a User model for restoration.
func UserInfoArchiveToUser(uInfo structs.UserInfoArchive) *user.User {
	u := &user.User{
		Username:  uInfo.Username,
		Password:  uInfo.Password,
		Activated: uInfo.Activated,
		UserPerms: user_model.Permissions(uInfo.PermissionLevel),
		HomeID:    uInfo.HomeID,
		TrashID:   uInfo.TrashID,
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
