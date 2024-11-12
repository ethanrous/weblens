package rest

import (
	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/models"
)

type BackupInfo struct {
	FileHistory    []*fileTree.Lifetime
	LifetimesCount int
	Users          []UserInfoArchive
	Instances      []ServerInfo
	ApiKeys        []models.ApiKey
}

func NewBackupInfo(fileHistory []*fileTree.Lifetime, lifetimesCount int, users []*models.User, instances []*models.Instance, apiKeys []models.ApiKey) BackupInfo {
	var userInfos []UserInfoArchive
	for _, u := range users {
		userInfos = append(userInfos, UserToUserInfoArchive(u))
	}

	var serverInfos []ServerInfo
	for _, i := range instances {
		serverInfos = append(serverInfos, InstanceToServerInfo(i))
	}

	return BackupInfo{
		FileHistory:    fileHistory,
		LifetimesCount: lifetimesCount,
		Users:          userInfos,
		Instances:      serverInfos,
		ApiKeys:        apiKeys,
	}
}
