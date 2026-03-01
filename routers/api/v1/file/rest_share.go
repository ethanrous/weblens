package file

import (
	"net/http"

	"github.com/ethanrous/weblens/models/db"
	share_model "github.com/ethanrous/weblens/models/share"
	user_model "github.com/ethanrous/weblens/models/usermodel"
	"github.com/ethanrous/weblens/modules/netwrk"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/modules/wlstructs"
	"github.com/ethanrous/weblens/services/ctxservice"
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
//	@Param		request	body		wlstructs.FileShareParams	true	"New File Share Params"
//	@Success	200		{object}	wlstructs.ShareInfo		"New File Share"
//	@Success	409
//	@Router		/share/file [post]
func CreateFileShare(ctx ctxservice.RequestContext) {
	shareParams, err := netwrk.ReadRequestBody[wlstructs.FileShareParams](ctx.Req)
	if err != nil {
		return
	}

	file, err := ctx.FileService.GetFileByID(ctx, shareParams.FileID)
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
		ctx.Error(http.StatusNotFound, wlerrors.New("you are not the owner of this file"))

		return
	}

	_, err = share_model.GetShareByFileID(ctx, shareParams.FileID)
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

// GetFileShare godoc
//
//	@ID			GetFileShare
//
//	@Summary	Get a file share
//	@Tags		Share
//	@Produce	json
//	@Param		shareID	path		string				true	"Share ID"
//	@Success	200		{object}	wlstructs.ShareInfo	"File Share"
//	@Failure	404
//	@Router		/share/{shareID} [get]
func GetFileShare(ctx ctxservice.RequestContext) {
	shareID := share_model.IDFromString(ctx.Path("shareID"))

	share, err := share_model.GetShareByID(ctx, shareID)
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
//	@Param		shareID	path	string	true	"Share ID"
//	@Param		public	query	bool	true	"Share Public Status"
//	@Success	200
//	@Failure	404
//	@Router		/share/{shareID}/public [patch]
func SetSharePublic(ctx ctxservice.RequestContext) {
	shareID := share_model.IDFromString(ctx.Path("shareID"))

	share, err := share_model.GetShareByID(ctx, shareID)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)

		return
	}

	publicStr := ctx.Query("public")
	if publicStr != "true" && publicStr != "false" {
		ctx.Error(http.StatusBadRequest, wlerrors.New("public query parameter must be 'true' or 'false'"))

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
//	@Param		shareID	path		string					true	"Share ID"
//	@Param		request	body		wlstructs.AddUserParams	true	"Share Accessors"
//	@Success	200		{object}	wlstructs.ShareInfo
//	@Failure	404
//	@Router		/share/{shareID}/accessors [post]
func AddUserToShare(ctx ctxservice.RequestContext) {
	shareID := share_model.IDFromString(ctx.Path("shareID"))

	share, err := share_model.GetShareByID(ctx, shareID)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)

		return
	}

	addUserBody, err := netwrk.ReadRequestBody[wlstructs.AddUserParams](ctx.Req)
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
		ctx.Error(http.StatusNotFound, wlerrors.New("user does not exist"))

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
//	@Param		shareID		path		string	true	"Share ID"
//	@Param		username	path		string	true	"Username"
//	@Success	200			{object}	wlstructs.ShareInfo
//	@Failure	404
//	@Router		/share/{shareID}/accessors/{username} [delete]
func RemoveUserFromShare(ctx ctxservice.RequestContext) {
	shareID := share_model.IDFromString(ctx.Path("shareID"))

	share, err := share_model.GetShareByID(ctx, shareID)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)

		return
	}

	username := ctx.Path("username")

	perms := share.GetUserPermissions(username)
	if perms == nil {
		ctx.Error(http.StatusNotFound, wlerrors.New("share does not include user"))

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

// SetShareAccessors godoc
//
//	@ID			UpdateShareAccessorPermissions
//
//	@Summary	Update a share's user permissions
//	@Tags		Share
//	@Produce	json
//	@Param		shareID		path		string						true	"Share ID"
//	@Param		username	path		string						true	"Username"
//	@Param		request		body		wlstructs.PermissionsParams	true	"Share Permissions Params"
//	@Success	200			{object}	wlstructs.ShareInfo
//	@Failure	404
//	@Router		/share/{shareID}/accessors/{username} [patch]
func SetShareAccessors(ctx ctxservice.RequestContext) {
	shareID := share_model.IDFromString(ctx.Path("shareID"))

	share, err := share_model.GetShareByID(ctx, shareID)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)

		return
	}

	permissionsBody, err := netwrk.ReadRequestBody[wlstructs.PermissionsParams](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	username := ctx.Path("username")

	existingPerms := share.GetUserPermissions(username)
	if existingPerms == nil {
		ctx.Error(http.StatusNotFound, wlerrors.New("user does not exist"))

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

// DeleteShare godoc
//
//	@ID			DeleteFileShare
//
//	@Summary	Delete a file share
//	@Tags		Share
//	@Produce	json
//	@Param		shareID	path	string	true	"Share ID"
//	@Success	200
//	@Failure	404
//	@Router		/share/{shareID} [delete]
func DeleteShare(ctx ctxservice.RequestContext) {
	shareID := share_model.IDFromString(ctx.Path("shareID"))

	err := share_model.DeleteShare(ctx, shareID)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}
}
