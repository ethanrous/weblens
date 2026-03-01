// Package auth provides authentication and authorization services for file and share access.
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
	user_model "github.com/ethanrous/weblens/models/usermodel"
	"github.com/ethanrous/weblens/modules/cryptography"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/modules/wlog"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
)

// ErrBadAuthHeader is returned when the Authorization header has an invalid format.
var ErrBadAuthHeader = wlerrors.Statusf(http.StatusBadRequest, "invalid auth header format")

// ErrMustAuthenticate is returned when authentication is required but not provided.
var ErrMustAuthenticate = wlerrors.Statusf(http.StatusUnauthorized, "user must authenticate to access this resource")

// ErrFileAccessNotPermitted is returned when a user lacks permission to access a file.
var ErrFileAccessNotPermitted = wlerrors.Statusf(http.StatusForbidden, "file access not permitted")

// ErrShareDoesNotPermitFile is returned when a share does not grant access to a specific file.
var ErrShareDoesNotPermitFile = wlerrors.Statusf(http.StatusForbidden, "share does not permit access to this file")

func doesSharePermitFile(_ context.Context, file *file_model.WeblensFileImpl, share *share_model.FileShare) bool {
	if share == nil || !share.Enabled || file.IsPastFile() {
		return false
	}

	for {
		if share.FileID == file.ID() {
			return true
		}

		file = file.GetParent()

		if file == nil {
			break
		}
	}

	return false
}

// CanUserAccessFile checks if a user has permission to access a file through a share.
func CanUserAccessFile(ctx context.Context, user *user_model.User, file *file_model.WeblensFileImpl, share *share_model.FileShare, requiredPerms ...share_model.Permission) (*share_model.Permissions, error) {
	if file.GetPortablePath() == file_model.UsersRootPath {
		if user.IsOwner() {
			return share_model.NewPermissions(), nil
		}

		return &share_model.Permissions{}, wlerrors.Statusf(http.StatusForbidden, "cannot access the USERS root path")
	}

	ownerName, err := file_model.GetFileOwnerName(ctx, file)
	if err != nil {
		wlog.FromContext(ctx).Error().Stack().Err(err).Msg("Failed to get file owner name")

		return &share_model.Permissions{}, err
	}

	// If the user is the owner of the file, we can access it regardless of the share
	// FIXME: Make admin access more granular. The current behavior is so backup operations can work.
	if ownerName == user.GetUsername() || user.IsAdmin() {
		return share_model.NewFullPermissions(), nil
	}

	// Check that the share permits access to the specific file we are trying to access
	if !doesSharePermitFile(ctx, file, share) {
		if share != nil {
			shareID := share.ShareID.Hex()

			return &share_model.Permissions{}, wlerrors.ReplaceStack(wlerrors.Errorf("denying user [%s] access to file [%s] using share [%s]: %w", user.Username, file.ID(), shareID, ErrShareDoesNotPermitFile))
		}

		return &share_model.Permissions{}, wlerrors.ReplaceStack(wlerrors.Errorf("denying user [%s] access to file [%s]: %w", user.Username, file.ID(), ErrFileAccessNotPermitted))
	}

	if user == nil || user.IsPublic() {
		if share != nil && share.IsPublic() {
			return share_model.NewPermissions(), nil
		}

		return &share_model.Permissions{}, ErrMustAuthenticate
	}

	if user.IsSystemUser() && user.Username == "WEBLENS" {
		return share_model.NewFullPermissions(), nil
	}

	allowedPerms := share.GetUserPermissions(user.GetUsername())
	if allowedPerms == nil && !share.Public {
		// If the user is not in the accessors list, we cannot access it
		return &share_model.Permissions{}, ErrFileAccessNotPermitted
	} else if allowedPerms != nil {
		for _, requiredPerm := range requiredPerms {
			if !allowedPerms.HasPermission(requiredPerm) {
				wlog.FromContext(ctx).Debug().Msgf("User [%s] does not have permission: %s", user.GetUsername(), requiredPerm)
				// If the user does not have the required permissions, we say cannot access it at all
				return &share_model.Permissions{}, ErrFileAccessNotPermitted
			}
		}

		// If the user has the required permissions, we can access it
		return allowedPerms, nil
	}
	// If the share is public, and allows access to the specific file we want, we can access it regardless of the accessors list
	return share.GetUserPermissions(user_model.PublicUserName), nil
}

// CanUserModifyShare checks if a user has permission to modify a share.
func CanUserModifyShare(user *user_model.User, share share_model.FileShare) bool {
	return user.GetUsername() == share.GetOwner()
}

// SetSessionToken generates and sets session cookies for the authenticated user.
func SetSessionToken(ctx context_service.RequestContext) error {
	if ctx.Requester == nil {
		return wlerrors.New("requester is nil")
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

// GenerateJWTCookie creates a session cookie containing a JWT for the user.
func GenerateJWTCookie(user *user_model.User) (string, error) {
	token, expires, err := cryptography.GenerateJWT(user.GetUsername())
	if err != nil {
		return "", err
	}

	cookie := fmt.Sprintf("%s=%s;Path=/;Expires=%s;HttpOnly", cryptography.SessionTokenCookie, token, expires.Format(time.RFC1123))

	return cookie, nil
}

// GenerateUserCookie creates a cookie containing the username.
func GenerateUserCookie(user *user_model.User) string {
	expires := time.Now().Add(time.Hour * 24 * 7).In(time.UTC)
	cookie := fmt.Sprintf("%s=%s;Path=/;Expires=%s;HttpOnly", cryptography.UserCrumbCookie, user.Username, expires.Format(time.RFC1123))

	return cookie
}

// GetUserFromJWT extracts and validates a user from a JWT token string.
func GetUserFromJWT(ctx context.Context, tokenStr string) (*user_model.User, error) {
	username, err := cryptography.GetUsernameFromToken(tokenStr)
	if err != nil {
		return nil, err
	}

	u, err := user_model.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// GetUserFromAuthHeader extracts and validates a user from an Authorization header.
func GetUserFromAuthHeader(ctx context.Context, authHeader string) (*user_model.User, error) {
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return nil, wlerrors.WrapStatus(http.StatusBadRequest, ErrBadAuthHeader)
	}

	var tokenStr string

	_, err := fmt.Sscanf(authHeader, "Bearer %s", &tokenStr)
	if err != nil {
		return nil, wlerrors.WrapStatus(http.StatusInternalServerError, err)
	}

	tokenByteSlice, err := base64.StdEncoding.DecodeString(tokenStr)
	if err != nil {
		return nil, wlerrors.WrapStatus(http.StatusInternalServerError, err)
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

// func (accSrv *AccessServiceImpl) SetKeyUsedBy(key models.WeblensAPIKey, remote *models.Instance) error {
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
// 	newUsingID := ""
// 	if remote != nil {
// 		newUsingID = remote.ServerID()
// 	}
//
// 	filter := bson.M{"key": key}
// 	update := bson.M{"$set": bson.M{"remoteUsing": newUsingID}}
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
