package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"slices"
	"time"

	auth_model "github.com/ethanrous/weblens/models/auth"
	file_model "github.com/ethanrous/weblens/models/file"
	share_model "github.com/ethanrous/weblens/models/share"
	user_model "github.com/ethanrous/weblens/models/user"
	context_mod "github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/crypto"
	"github.com/ethanrous/weblens/modules/errors"
)

var ErrBadAuthHeader = errors.New("invalid auth header format")

func doesSharePermitFile(ctx context.Context, file *file_model.WeblensFileImpl, share *share_model.FileShare) bool {
	if share == nil || !share.Enabled || file.IsPastFile() {
		return false
	}

	if share.FileId == file.ID() {
		return true
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

func CanUserAccessFile(ctx context.Context, user *user_model.User, file *file_model.WeblensFileImpl, share *share_model.FileShare) bool {
	sharePermitsFile := doesSharePermitFile(ctx, file, share)

	if user == nil || user.IsPublic() {
		return share != nil && share.IsPublic() && sharePermitsFile
	}

	ownerName, err := file_model.GetFileOwnerName(ctx, file)
	if err != nil {
		context_mod.ToZ(ctx).Log().Error().Stack().Err(err).Msg("Failed to get file owner name")

		return false
	}

	if ownerName == user.GetUsername() {
		return true
	}

	if user.IsSystemUser() && user.Username == "WEBLENS" {
		return true
	}

	if !sharePermitsFile {
		// If the share does not permit access to the file, we cannot access it now we know
		// the user is not the owner of the file
		return false
	} else if share.Public {
		// If the share is public, and allows access to the specific file we want, we can access it regardless of the accessors list
		return true
	}

	// Share is now not public, but the share does allow access to the file. We need to check if the user is in the accessors list.
	if slices.Contains(share.Accessors, user.GetUsername()) {
		return true
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

func GetUserFromAuthHeader(ctx context.Context, authHeader string) (*user_model.User, error) {
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return nil, errors.WrapStatus(http.StatusBadRequest, ErrBadAuthHeader)
	}

	_, err := fmt.Sscanf(authHeader, "Bearer %s", &authHeader)
	if err != nil {
		return nil, errors.WrapStatus(http.StatusInternalServerError, err)
	}

	tokenByteSlice, err := base64.StdEncoding.DecodeString(authHeader)
	if err != nil {
		return nil, errors.WrapStatus(http.StatusInternalServerError, err)
	}

	var tokenBytes [32]byte

	copy(tokenBytes[:], tokenByteSlice)

	token, err := auth_model.GetToken(ctx, tokenBytes)
	if err != nil {
		return nil, err
	}

	u, err := user_model.GetUserByUsername(ctx, token.Owner)
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
