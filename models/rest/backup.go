package rest

import (
	"context"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/models/auth"
	"github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/structs"
	"github.com/ethanrous/weblens/services/reshape"
)

type BackupInfo struct {
	FileHistory    []*fileTree.Lifetime
	Users          []structs.UserInfoArchive
	Instances      []structs.TowerInfo
	Tokens         []structs.Token
	LifetimesCount int
}

func NewBackupInfo(ctx context.Context, fileHistory []*fileTree.Lifetime, lifetimesCount int, users []*user.User, instances []*tower.Instance, tokens []*auth.Token) BackupInfo {
	var userInfos []structs.UserInfoArchive
	for _, u := range users {
		userInfos = append(userInfos, reshape.UserToUserInfoArchive(u))
	}

	var serverInfos []structs.TowerInfo
	for _, i := range instances {
		serverInfos = append(serverInfos, reshape.TowerToTowerInfo(i))
	}

	var tokenInfos []structs.Token
	for _, k := range tokens {
		tokenInfos = append(tokenInfos, reshape.TokenToTokenInfo(k))
	}

	return BackupInfo{
		FileHistory:    fileHistory,
		LifetimesCount: lifetimesCount,
		Users:          userInfos,
		Instances:      serverInfos,
		Tokens:         tokenInfos,
	}
}
