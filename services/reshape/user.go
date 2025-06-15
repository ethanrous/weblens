package reshape

import (
	"context"

	openapi "github.com/ethanrous/weblens/api"
	"github.com/ethanrous/weblens/models/user"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/structs"
)

func UserToUserInfo(ctx context.Context, u *user_model.User) structs.UserInfo {
	return structs.UserInfo{
		Username:        u.Username,
		FullName:        u.DisplayName,
		HomeId:          u.HomeId,
		TrashId:         u.TrashId,
		PermissionLevel: int(u.UserPerms),
		Activated:       u.Activated,
	}
}

func UserToUserInfoArchive(u *user.User) structs.UserInfoArchive {
	if u == nil || u.IsSystemUser() {
		return structs.UserInfoArchive{}
	}

	info := structs.UserInfoArchive{
		Password:  u.Password,
		Activated: u.IsActive(),
	}
	info.Username = u.GetUsername()
	info.FullName = u.DisplayName
	info.PermissionLevel = int(u.UserPerms)
	info.HomeId = u.HomeId
	info.TrashId = u.TrashId

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
