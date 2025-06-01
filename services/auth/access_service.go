package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	auth_model "github.com/ethanrous/weblens/models/auth"
	file_model "github.com/ethanrous/weblens/models/file"
	share_model "github.com/ethanrous/weblens/models/share"
	user_model "github.com/ethanrous/weblens/models/user"
	context_mod "github.com/ethanrous/weblens/modules/context"
	"github.com/ethanrous/weblens/modules/crypto"
	"github.com/ethanrous/weblens/modules/errors"
	context_service "github.com/ethanrous/weblens/services/context"
)

var ErrBadAuthHeader = errors.Statusf(http.StatusBadRequest, "invalid auth header format")
var ErrMustAuthenticate = errors.Statusf(http.StatusUnauthorized, "user must authenticate to access this resource")
var ErrFileAccessNotPermitted = errors.Statusf(http.StatusForbidden, "file access not permitted")
var ErrShareDoesNotPermitFile = errors.Statusf(http.StatusForbidden, "share does not permit access to this file")

func doesSharePermitFile(ctx context.Context, file *file_model.WeblensFileImpl, share *share_model.FileShare) bool {
	if share == nil || !share.Enabled || file.IsPastFile() {
		return false
	}

	for {
		if share.FileId == file.ID() {
			return true
		}

		file = file.GetParent()

		if file == nil {
			break
		}
	}

	return false
}

func CanUserAccessFile(ctx context.Context, user *user_model.User, file *file_model.WeblensFileImpl, share *share_model.FileShare, requiredPerms ...share_model.Permission) error {
	ownerName, err := file_model.GetFileOwnerName(ctx, file)
	if err != nil {
		context_mod.ToZ(ctx).Log().Error().Stack().Err(err).Msg("Failed to get file owner name")

		return err
	}

	// If the user is the owner of the file, we can access it regardless of the share
	if ownerName == user.GetUsername() {
		return nil
	}

	sharePermitsFile := doesSharePermitFile(ctx, file, share)
	if !sharePermitsFile && share != nil && share.Enabled {
		return errors.Errorf("invalid share [%s] for file [%s]: %w", share.ShareId, file.ID(), ErrShareDoesNotPermitFile)
	}

	if user == nil || user.IsPublic() {
		if share != nil && share.IsPublic() && sharePermitsFile {
			return nil
		} else {
			return ErrMustAuthenticate
		}
	}

	if user.IsSystemUser() && user.Username == "WEBLENS" {
		return nil
	}

	if !sharePermitsFile {
		// If the share does not permit access to the file, we cannot access it now we know
		// the user is not the owner of the file
		return ErrFileAccessNotPermitted
	} else if share.Public {
		// If the share is public, and allows access to the specific file we want, we can access it regardless of the accessors list
		return nil
	}

	appCtx, _ := context_service.FromContext(ctx)
	appCtx.Log().Debug().Msgf("Checking file access for user [%s] on file [%s]", user.GetUsername(), file.ID())

	allowedPerms := share.GetUserPermissions(user.GetUsername())
	if allowedPerms == nil {
		// If the user is not in the accessors list, we cannot access it
		return ErrFileAccessNotPermitted
	}

	for _, requiredPerm := range requiredPerms {
		if !allowedPerms.HasPermission(requiredPerm) {
			appCtx.Log().Debug().Msgf("User [%s] does not have permission [%s] on file [%s]", user.GetUsername(), requiredPerm, file.GetPortablePath().String())

			return ErrFileAccessNotPermitted
		}
	}

	return nil
}

func CanUserModifyShare(user *user_model.User, share share_model.FileShare) bool {
	return user.GetUsername() == share.GetOwner()
}

func SetSessionToken(ctx context_service.RequestContext) error {
	if ctx.Requester == nil {
		return errors.New("requester is nil")
	}

	cookie, err := GenerateJWTCookie(ctx.Requester)
	if err != nil {
		return err
	}

	ctx.SetHeader("Set-Cookie", cookie)

	return nil
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
