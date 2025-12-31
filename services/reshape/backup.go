// Package reshape provides functions for converting between domain models and API transfer objects.
package reshape

import (
	"context"

	"github.com/ethanrous/weblens/models/auth"
	"github.com/ethanrous/weblens/models/history"
	"github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/structs"
)

// NewBackupInfo creates a backup information object from file history, users, instances, and tokens.
func NewBackupInfo(ctx context.Context, fileHistory []history.FileAction, users []*user.User, instances []tower.Instance, tokens []*auth.Token) structs.BackupInfo {
	fileActionInfos := make([]structs.FileActionInfo, 0, len(fileHistory))
	for _, a := range fileHistory {
		fileActionInfos = append(fileActionInfos, FileActionToFileActionInfo(a))
	}

	userInfos := make([]structs.UserInfoArchive, 0, len(users))
	for _, u := range users {
		userInfos = append(userInfos, UserToUserInfoArchive(ctx, u))
	}

	serverInfos := make([]structs.TowerInfo, 0, len(instances))
	for _, i := range instances {
		serverInfos = append(serverInfos, TowerToTowerInfo(ctx, i))
	}

	tokenInfos := make([]structs.TokenInfo, 0, len(tokens))
	for _, k := range tokens {
		tokenInfos = append(tokenInfos, TokenToTokenInfo(ctx, k))
	}

	return structs.BackupInfo{
		FileHistory: fileActionInfos,
		Users:       userInfos,
		Instances:   serverInfos,
		Tokens:      tokenInfos,
	}
}
