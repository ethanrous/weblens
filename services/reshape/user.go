package reshape

import (
	"context"

	openapi "github.com/ethanrous/weblens/api"
	"github.com/ethanrous/weblens/models/user"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/structs"
	context_service "github.com/ethanrous/weblens/services/context"
)

func UserToUserInfo(ctx context.Context, u *user_model.User) structs.UserInfo {
	userIsOnline := getUserIsOnline(ctx, u.Username)

	return structs.UserInfo{
		Username:        u.Username,
		FullName:        u.DisplayName,
		HomeId:          u.HomeId,
		TrashId:         u.TrashId,
		PermissionLevel: int(u.UserPerms),
		Activated:       u.Activated,
		IsOnline:        userIsOnline,
	}
}

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
			HomeId:          u.HomeId,
			TrashId:         u.TrashId,
			Activated:       u.IsActive(),
			IsOnline:        userIsOnline,
		},
		Password: u.Password,
	}

	return info
}

func UserInfoArchiveToUser(uInfo openapi.UserInfoArchive) *user.User {
	u := &user.User{
		Username:  uInfo.Username,
		Password:  uInfo.GetPassword(),
		Activated: uInfo.Activated,
		UserPerms: user_model.UserPermissions(uInfo.PermissionLevel),
		HomeId:    uInfo.HomeId,
		TrashId:   uInfo.TrashId,
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
