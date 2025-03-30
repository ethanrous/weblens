package auth

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	share_model "github.com/ethanrous/weblens/models/share"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/crypto"
)

func CanUserAccessFile(
	user *user_model.User, file *file_model.WeblensFileImpl, share *share_model.FileShare,
) bool {
	if user == nil || user.IsPublic() {
		return share != nil && share.IsPublic()
	}

	if getFileOwnerName(file) == user.GetUsername() {
		return true
	}

	if user.IsSystemUser() && user.Username == "WEBLENS" {
		return true
	}

	if share == nil || !share.Enabled || (!share.Public && !slices.Contains(share.Accessors, user.GetUsername())) {
		return false
	}

	tmpFile := file
	for tmpFile.GetParent() != nil {
		if tmpFile.ID() == share.FileId {
			return true
		}
		tmpFile = tmpFile.GetParent()
	}
	return false
}

func CanUserModifyShare(user *user_model.User, share share_model.FileShare) bool {
	return user.GetUsername() == share.GetOwner()
}

func GenerateJWTCookie(user *user_model.User) (string, error) {
	token, expires, err := crypto.GenerateJWT(user.GetUsername())
	if err != nil {
		return "", err
	}

	cookie := fmt.Sprintf("%s=%s;Path=/;Expires=%s;HttpOnly", crypto.SessionTokenCookie, token, expires.Format(time.RFC1123))
	return cookie, nil

}

func GetUserFromJWT(ctx context.Context, tokenStr string) (*user_model.User, error) {
	username, err := crypto.GetUsernameFromToken(tokenStr)
	if err != nil {
		return nil, err
	}

	u, err := user_model.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// func (accSrv *AccessServiceImpl) SetKeyUsedBy(key models.WeblensApiKey, remote *models.Instance) error {
// 	accSrv.keyMapMu.RLock()
// 	keyInfo, ok := accSrv.apiKeyMap[key]
// 	accSrv.keyMapMu.RUnlock()
//
// 	if !ok {
// 		return werror.WithStack(werror.ErrKeyNotFound)
// 	}
//
// 	if keyInfo.RemoteUsing != "" && remote != nil {
// 		return werror.WithStack(werror.ErrKeyInUse)
// 	}
//
// 	newUsingId := ""
// 	if remote != nil {
// 		newUsingId = remote.ServerId()
// 	}
//
// 	filter := bson.M{"key": key}
// 	update := bson.M{"$set": bson.M{"remoteUsing": newUsingId}}
// 	_, err := accSrv.collection.UpdateOne(context.Background(), filter, update)
// 	if err != nil {
// 		return werror.WithStack(err)
// 	}
//
// 	keyInfo.RemoteUsing = newUsingId
// 	accSrv.keyMapMu.Lock()
// 	accSrv.apiKeyMap[key] = keyInfo
// 	accSrv.keyMapMu.Unlock()
//
// 	return nil
// }

func getFileOwnerName(file *file_model.WeblensFileImpl) string {
	portable := file.GetPortablePath()
	if portable.RootName() != "USERS" {
		return "WEBLENS"
	}
	slashIndex := strings.Index(portable.RelativePath(), "/")
	var username string
	if slashIndex == -1 {
		username = string(portable.RelativePath())
	} else {
		username = string(portable.RelativePath()[:slashIndex])
	}

	return username
}
