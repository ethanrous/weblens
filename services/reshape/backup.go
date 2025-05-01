package reshape

import (
	"context"

	"github.com/ethanrous/weblens/models/auth"
	"github.com/ethanrous/weblens/models/history"
	"github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/structs"
)

func NewBackupInfo(ctx context.Context, fileHistory []history.FileAction, users []*user.User, instances []tower.Instance, tokens []*auth.Token) structs.BackupInfo {
	var fileActionInfos []structs.FileActionInfo
	for _, a := range fileHistory {
		fileActionInfos = append(fileActionInfos, FileActionToFileActionInfo(a))
	}

	var userInfos []structs.UserInfoArchive
	for _, u := range users {
		userInfos = append(userInfos, UserToUserInfoArchive(u))
	}

	var serverInfos []structs.TowerInfo
	for _, i := range instances {
		serverInfos = append(serverInfos, TowerToTowerInfo(ctx, i))
	}

	var tokenInfos []structs.TokenInfo
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
