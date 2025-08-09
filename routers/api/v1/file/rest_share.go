package file

import (
	"net/http"

	"github.com/ethanrous/weblens/models/db"
	share_model "github.com/ethanrous/weblens/models/share"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/net"
	"github.com/ethanrous/weblens/modules/structs"
	"github.com/ethanrous/weblens/services/context"
	file_service "github.com/ethanrous/weblens/services/file"
	"github.com/ethanrous/weblens/services/reshape"
)

// CreateFileShare godoc
//
//	@ID			CreateFileShare
//
//	@Summary	Share a file
//	@Tags		Share
//	@Produce	json
//	@Param		request	body		structs.FileShareParams	true	"New File Share Params"
//	@Success	200		{object}	structs.ShareInfo		"New File Share"
//	@Success	409
//	@Router		/share/file [post]
func CreateFileShare(ctx context.RequestContext) {
	shareParams, err := net.ReadRequestBody[structs.FileShareParams](ctx.Req)
	if err != nil {
		return
	}

	file, err := ctx.FileService.GetFileById(ctx, shareParams.FileId)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)

		return
	}

	owner, err := file_service.GetFileOwner(ctx, file)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	if owner.GetUsername() != ctx.Requester.GetUsername() {
		// TODO: Obfuscate error message so that it doesn't leak information about existing files
		ctx.Error(http.StatusNotFound, errors.New("you are not the owner of this file"))

		return
	}

	_, err = share_model.GetShareByFileId(ctx, shareParams.FileId)
	if err == nil {
		ctx.Error(http.StatusConflict, share_model.ErrShareAlreadyExists)

		return
	}

	if !db.IsNotFound(err) {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	accessors := make([]*user_model.User, 0, len(shareParams.Users))

	for _, un := range shareParams.Users {
		u, err := user_model.GetUserByUsername(ctx, un)
		if err != nil {
			ctx.Error(http.StatusNotFound, err)

			return
		}

		accessors = append(accessors, u)
	}

	newShare, err := share_model.NewFileShare(ctx, file.ID(), ctx.Requester, accessors, shareParams.Public, shareParams.Wormhole, shareParams.TimelineOnly)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	err = share_model.SaveFileShare(ctx, newShare)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	newShareInfo := reshape.ShareToShareInfo(ctx, newShare)
	ctx.JSON(http.StatusCreated, newShareInfo)
}

// CreateAlbumShare godoc
//
//	@ID			CreateAlbumShare
//
//	@Summary	Share an album
//	@Tags		Share
//	@Produce	json
//	@Param		request	body		structs.AlbumShareParams	true	"New Album Share Params"
//	@Success	200		{object}	structs.ShareInfo			"New Album Share"
//	@Success	409
//	@Router		/share/album [post]
// func CreateAlbumShare(ctx context.RequestContext) {
// 	shareParams, err := net.ReadRequestBody[structs.AlbumShareParams](ctx.Req)
// 	if err != nil {
// 		return
// 	}
//
// 	album := pack.AlbumService.Get(shareParams.AlbumId)
// 	if album == nil {
// 		SafeErrorAndExit(werror.ErrNoAlbum, w)
// 		return
// 	}
//
// 	_, err = pack.ShareService.GetAlbumShare(album.ID())
// 	if !errors.Is(err, werror.ErrNoShare) {
// 		SafeErrorAndExit(werror.ErrShareAlreadyExists, w)
// 		return
// 	} else if SafeErrorAndExit(err, w) {
// 		return
// 	}
//
// 	accessors := internal.Map(
// 		shareParams.Users, func(un string) *models.User {
// 			return pack.UserService.Get(un)
// 		},
// 	)
//
// 	newShare := models.NewAlbumShare(album, u, accessors, shareParams.Public)
//
// 	err = pack.ShareService.Add(newShare)
// 	if SafeErrorAndExit(err, w) {
// 		return
// 	}
//
// 	writeJson(w, http.StatusCreated, newShare)
// }

// GetFileShare godoc
//
//	@ID			GetFileShare
//
//	@Summary	Get a file share
//	@Tags		Share
//	@Produce	json
//	@Param		shareId	path		string				true	"Share Id"
//	@Success	200		{object}	structs.ShareInfo	"File Share"
//	@Failure	404
//	@Router		/share/{shareId} [get]
func GetFileShare(ctx context.RequestContext) {
	shareId := share_model.ShareIdFromString(ctx.Path("shareId"))
	share, err := share_model.GetShareById(ctx, shareId)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)

		return
	}

	// TODO: check permissions

	shareInfo := reshape.ShareToShareInfo(ctx, share)
	ctx.JSON(http.StatusOK, shareInfo)
}

// SetSharePublic godoc
//
//	@ID			SetSharePublic
//
//	@Summary	Update a share's "public" status
//	@Tags		Share
//	@Produce	json
//	@Param		shareId	path	string	true	"Share Id"
//	@Param		public	query	bool	true	"Share Public Status"
//	@Success	200
//	@Failure	404
//	@Router		/share/{shareId}/public [patch]
func SetSharePublic(ctx context.RequestContext) {
	shareId := share_model.ShareIdFromString(ctx.Path("shareId"))

	share, err := share_model.GetShareById(ctx, shareId)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)

		return
	}

	publicStr := ctx.Query("public")
	if publicStr != "true" && publicStr != "false" {
		ctx.Error(http.StatusBadRequest, errors.New("public query parameter must be 'true' or 'false'"))

		return
	}

	err = share.SetPublic(ctx, publicStr == "true")
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.Status(http.StatusOK)
}

// AddUserToShare godoc
//
//	@ID			AddUserToShare
//
//	@Summary	Add a user to a file share
//	@Tags		Share
//	@Produce	json
//	@Param		shareId	path		string					true	"Share Id"
//	@Param		request	body		structs.AddUserParams	true	"Share Accessors"
//	@Success	200		{object}	structs.ShareInfo
//	@Failure	404
//	@Router		/share/{shareId}/accessors [post]
func AddUserToShare(ctx context.RequestContext) {
	shareId := share_model.ShareIdFromString(ctx.Path("shareId"))

	share, err := share_model.GetShareById(ctx, shareId)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)

		return
	}

	addUserBody, err := net.ReadRequestBody[structs.AddUserParams](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	newUsername, params, err := reshape.UnpackNewUserParams(ctx, addUserBody)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	exists, err := user_model.DoesUserExist(ctx, newUsername)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	} else if !exists {
		ctx.Error(http.StatusNotFound, errors.New("user does not exist"))

		return
	}

	err = share.AddUser(ctx, newUsername, &params)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	shareInfo := reshape.ShareToShareInfo(ctx, share)
	ctx.JSON(http.StatusOK, shareInfo)
}

// RemoveUserFromShare godoc
//
//	@ID			RemoveUserFromShare
//
//	@Summary	Remove a user from a file share
//	@Tags		Share
//	@Produce	json
//	@Param		shareId		path		string	true	"Share Id"
//	@Param		username	path		string	true	"Username"
//	@Success	200			{object}	structs.ShareInfo
//	@Failure	404
//	@Router		/share/{shareId}/accessors/{username} [delete]
func RemoveUserFromShare(ctx context.RequestContext) {
	shareId := share_model.ShareIdFromString(ctx.Path("shareId"))

	share, err := share_model.GetShareById(ctx, shareId)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)

		return
	}

	username := ctx.Path("username")

	perms := share.GetUserPermissions(username)
	if perms == nil {
		ctx.Error(http.StatusNotFound, errors.New("share does not include user"))

		return
	}

	err = share.RemoveUsers(ctx, []string{username})
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	shareInfo := reshape.ShareToShareInfo(ctx, share)
	ctx.JSON(http.StatusOK, shareInfo)
}

// UpdateShareAccessorPermissions godoc
//
//	@ID			UpdateShareAccessorPermissions
//
//	@Summary	Update a share's user permissions
//	@Tags		Share
//	@Produce	json
//	@Param		shareId		path		string						true	"Share Id"
//	@Param		username	path		string						true	"Username"
//	@Param		request		body		structs.PermissionsParams	true	"Share Permissions Params"
//	@Success	200			{object}	structs.ShareInfo
//	@Failure	404
//	@Router		/share/{shareId}/accessors/{username} [patch]
func SetShareAccessors(ctx context.RequestContext) {
	shareId := share_model.ShareIdFromString(ctx.Path("shareId"))

	share, err := share_model.GetShareById(ctx, shareId)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)

		return
	}

	permissionsBody, err := net.ReadRequestBody[structs.PermissionsParams](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	username := ctx.Path("username")

	existingPerms := share.GetUserPermissions(username)
	if existingPerms == nil {
		ctx.Error(http.StatusNotFound, errors.New("user does not exist"))

		return
	}

	perms, err := reshape.PermissionsParamsToPermissions(ctx, permissionsBody)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	if perms.CanDelete && !existingPerms.CanEdit {
		perms.CanEdit = true
	} else if !perms.CanEdit && existingPerms.CanDelete {
		perms.CanDelete = false
	}

	err = share.SetUserPermissions(ctx, username, &perms)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	shareInfo := reshape.ShareToShareInfo(ctx, share)
	ctx.JSON(http.StatusOK, shareInfo)
}

// DeleteFileShare godoc
//
//	@ID			DeleteFileShare
//
//	@Summary	Delete a file share
//	@Tags		Share
//	@Produce	json
//	@Param		shareId	path	string	true	"Share Id"
//	@Success	200
//	@Failure	404
//	@Router		/share/{shareId} [delete]
func DeleteShare(ctx context.RequestContext) {
	shareId := share_model.ShareIdFromString(ctx.Path("shareId"))

	err := share_model.DeleteShare(ctx, shareId)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}
}
