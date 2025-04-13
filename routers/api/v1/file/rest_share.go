package file

import (
	"net/http"

	share_model "github.com/ethanrous/weblens/models/share"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/net"
	"github.com/ethanrous/weblens/modules/structs"
	"github.com/ethanrous/weblens/services/context"
	file_service "github.com/ethanrous/weblens/services/file"
	"github.com/ethanrous/weblens/services/reshape"
	"github.com/pkg/errors"
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
func CreateFileShare(ctx *context.RequestContext) {
	shareInfo, err := net.ReadRequestBody[structs.FileShareParams](ctx.Req)
	if err != nil {
		return
	}

	file, err := ctx.FileService.GetFileById(shareInfo.FileId)
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

	_, err = share_model.GetShareByFileId(ctx, shareInfo.FileId)
	if !errors.Is(err, share_model.ErrShareAlreadyExists) {
		ctx.Error(http.StatusConflict, err)
		return
	} else if err != nil {
		ctx.Error(http.StatusInternalServerError, err)
		return
	}

	var accessors []*user_model.User
	for _, un := range shareInfo.Users {
		u, err := user_model.GetUserByUsername(ctx, un)
		if err != nil {
			ctx.Error(http.StatusNotFound, err)
			return
		}
		accessors = append(accessors, u)
	}

	newShare, err := share_model.NewFileShare(ctx, file.ID(), ctx.Requester, accessors, shareInfo.Public, shareInfo.Wormhole)
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
// func CreateAlbumShare(ctx *context.RequestContext) {
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
func GetFileShare(ctx *context.RequestContext) {
	shareId := ctx.Path("shareId")
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
func SetSharePublic(ctx *context.RequestContext) {
	shareId := ctx.Path("shareId")
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

// SetShareAccessors godoc
//
//	@ID			SetShareAccessors
//
//	@Summary	Update a share's accessors list
//	@Tags		Share
//	@Produce	json
//	@Param		shareId	path		string					true	"Share Id"
//	@Param		request	body		structs.UserListBody	true	"Share Accessors"
//	@Success	200		{object}	structs.ShareInfo
//	@Failure	404
//	@Router		/share/{shareId}/accessors [patch]
func SetShareAccessors(ctx *context.RequestContext) {
	shareId := ctx.Path("shareId")
	share, err := share_model.GetShareById(ctx, shareId)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)
		return
	}

	usersBody, err := net.ReadRequestBody[structs.UserListBody](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)
		return
	}

	var addUsers []string
	for _, un := range usersBody.AddUsers {
		u, err := user_model.GetUserByUsername(ctx, un)
		if err != nil {
			ctx.Error(http.StatusNotFound, err)
			return
		}
		addUsers = append(addUsers, u.Username)
	}

	var removeUsers []string
	for _, un := range usersBody.RemoveUsers {
		u, err := user_model.GetUserByUsername(ctx, un)
		if err != nil {
			ctx.Error(http.StatusNotFound, err)
			return
		}
		removeUsers = append(removeUsers, u.Username)
	}

	if len(addUsers) > 0 {
		err = share.AddUsers(ctx, addUsers)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)
			return
		}
	}

	if len(removeUsers) > 0 {
		err = share.RemoveUsers(ctx, addUsers)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)
			return
		}
	}

	share, err = share_model.GetShareById(ctx, shareId)
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
func DeleteShare(ctx *context.RequestContext) {
	shareId := ctx.Path("shareId")
	err := share_model.DeleteShare(ctx, shareId)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)
		return
	}
}
