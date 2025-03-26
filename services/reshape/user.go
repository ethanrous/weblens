package reshape

import (
	"context"

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

		// TODO: Add these fields
		// HomeSize:  u.HomeSize,
		// TrashSize: u.TrashSize,
	}
}

func UserToUserInfoArchive(u *models.User) UserInfoArchive {
	if u == nil || u.IsSystemUser() {
		return UserInfoArchive{}
	}
	info := UserInfoArchive{
		Password:  u.PasswordHash,
		Activated: u.IsActive(),
	}
	info.Username = u.GetUsername()
	info.FullName = u.GetFullName()
	info.Admin = u.IsAdmin()
	info.Owner = u.IsOwner()
	info.HomeId = u.HomeId
	info.TrashId = u.TrashId

	return info
}

func UserInfoArchiveToUser(uInfo UserInfoArchive) *models.User {
	u := &models.User{
		Username:      uInfo.Username,
		PasswordHash:  uInfo.Password,
		Activated:     uInfo.Activated,
		Admin:         uInfo.Admin,
		IsServerOwner: uInfo.Owner,
		HomeId:        uInfo.HomeId,
		TrashId:       uInfo.TrashId,
	}

	return u
}
