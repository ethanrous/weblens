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
	"github.com/ethanrous/weblens/modules/crypto"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/log"
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

func CanUserAccessFile(ctx context.Context, user *user_model.User, file *file_model.WeblensFileImpl, share *share_model.FileShare, requiredPerms ...share_model.Permission) (*share_model.Permissions, error) {
	ownerName, err := file_model.GetFileOwnerName(ctx, file)
	if err != nil {
		log.FromContext(ctx).Error().Stack().Err(err).Msg("Failed to get file owner name")

		return &share_model.Permissions{}, err
	}

	// If the user is the owner of the file, we can access it regardless of the share
	if ownerName == user.GetUsername() {
		return share_model.NewFullPermissions(), nil
	}

	// Check that the share permits access the specific file we are trying to access
	if !doesSharePermitFile(ctx, file, share) {
		shareId := ""
		if share != nil {
			shareId = share.ShareId.Hex()
		}

		return &share_model.Permissions{}, errors.Errorf("invalid share [%s] for file [%s]: %w", shareId, file.ID(), ErrShareDoesNotPermitFile)
	}

	if user == nil || user.IsPublic() {
		if share != nil && share.IsPublic() {
			return share_model.NewPermissions(), nil
		} else {
			return &share_model.Permissions{}, ErrMustAuthenticate
		}
	}

	if user.IsSystemUser() && user.Username == "WEBLENS" {
		return share_model.NewFullPermissions(), nil
	}

	if allowedPerms := share.GetUserPermissions(user.GetUsername()); allowedPerms == nil && !share.Public {
		// If the user is not in the accessors list, we cannot access it
		return &share_model.Permissions{}, ErrFileAccessNotPermitted
	} else if allowedPerms != nil {
		for _, requiredPerm := range requiredPerms {
			if !allowedPerms.HasPermission(requiredPerm) {
				log.FromContext(ctx).Debug().Msgf("User [%s] does not have permission: %s", user.GetUsername(), requiredPerm)
				// If the user does not have the required permissions, we say cannot access it at all
				return &share_model.Permissions{}, ErrFileAccessNotPermitted
			}
		}

		// If the user has the required permissions, we can access it
		return allowedPerms, nil
	} else {
		// If the share is public, and allows access to the specific file we want, we can access it regardless of the accessors list
		return share.GetUserPermissions(user_model.PublicUserName), nil
	}
}

func CanUserModifyShare(user *user_model.User, share share_model.FileShare) bool {
	return user.GetUsername() == share.GetOwner()
}

func SetSessionToken(ctx context_service.RequestContext) error {
	if ctx.Requester == nil {
		return errors.New("requester is nil")
	}

	sessionCookie, err := GenerateJWTCookie(ctx.Requester)
	if err != nil {
		return err
	}

	ctx.SetHeader("Set-Cookie", sessionCookie)

	usernameCookie := GenerateUserCookie(ctx.Requester)
	ctx.AddHeader("Set-Cookie", usernameCookie)

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

func GenerateUserCookie(user *user_model.User) string {
	expires := time.Now().Add(time.Hour * 24 * 7).In(time.UTC)
	cookie := fmt.Sprintf("%s=%s;Path=/;Expires=%s;HttpOnly", crypto.UserCrumbCookie, user.Username, expires.Format(time.RFC1123))

	return cookie
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
