package http

import (
	"errors"
	"net/http"

	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/rest"
	"github.com/go-chi/chi/v5"
)

// CreateFileShare godoc
//
//	@ID			CreateFileShare
//
//	@Summary	Share a file
//	@Tags		Share
//	@Produce	json
//	@Param		request	body		rest.FileShareParams	true	"New File Share Params"
//	@Success	200		{object}	rest.ShareInfo			"New File Share"
//	@Success	409
//	@Router		/share/file [post]
func createFileShare(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r, true)
	if SafeErrorAndExit(err, w) {
		return
	}

	shareInfo, err := readCtxBody[rest.FileShareParams](w, r)
	if err != nil {
		return
	}

	f, err := pack.FileService.GetFileSafe(shareInfo.FileId, u, nil)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = pack.ShareService.GetFileShare(f.ID())
	if err == nil {
		w.WriteHeader(http.StatusConflict)
		return
	} else if !errors.Is(err, werror.ErrNoShare) {
		log.ErrTrace(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	accessors := internal.Map(
		shareInfo.Users, func(un models.Username) *models.User {
			return pack.UserService.Get(un)
		},
	)
	newShare := models.NewFileShare(f, u, accessors, shareInfo.Public, shareInfo.Wormhole)

	err = pack.ShareService.Add(newShare)
	if err != nil {
		log.ShowErr(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	newShareInfo := rest.ShareToShareInfo(newShare)

	writeJson(w, http.StatusCreated, newShareInfo)
}

// CreateAlbumShare godoc
//
//	@ID			CreateAlbumShare
//
//	@Summary	Share an album
//	@Tags		Share
//	@Produce	json
//	@Param		request	body		rest.AlbumShareParams	true	"New Album Share Params"
//	@Success	200		{object}	rest.ShareInfo			"New Album Share"
//	@Success	409
//	@Router		/share/album [post]
func createAlbumShare(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r, true)
	if SafeErrorAndExit(err, w) {
		return
	}

	shareParams, err := readCtxBody[rest.AlbumShareParams](w, r)
	if err != nil {
		return
	}

	album := pack.AlbumService.Get(shareParams.AlbumId)
	if album == nil {
		SafeErrorAndExit(werror.ErrNoAlbum, w)
		return
	}

	_, err = pack.ShareService.GetAlbumShare(album.ID())
	if !errors.Is(err, werror.ErrNoShare) {
		SafeErrorAndExit(werror.ErrShareAlreadyExists, w)
		return
	} else if SafeErrorAndExit(err, w) {
		return
	}

	accessors := internal.Map(
		shareParams.Users, func(un models.Username) *models.User {
			return pack.UserService.Get(un)
		},
	)

	newShare := models.NewAlbumShare(album, u, accessors, shareParams.Public)

	err = pack.ShareService.Add(newShare)
	if SafeErrorAndExit(err, w) {
		return
	}

	writeJson(w, http.StatusCreated, newShare)
}

// GetFileShare godoc
//
//	@ID			GetFileShare
//
//	@Summary	Get a file share
//	@Tags		Share
//	@Produce	json
//	@Param		shareId	path		string			true	"Share Id"
//	@Success	200		{object}	rest.ShareInfo	"File Share"
//	@Failure	404
//	@Router		/share/{shareId} [get]
func getFileShare(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r, true)
	if SafeErrorAndExit(err, w) {
		return
	}
	shareId := models.ShareId(chi.URLParam(r, "shareId"))

	share := pack.ShareService.Get(shareId)
	if share == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fileShare, ok := share.(*models.FileShare)
	if !ok {
		log.Warning.Printf(
			"%s tried to get share [%s] as a fileShare (is %s)", u.GetUsername(), shareId, share.GetShareType(),
		)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	shareInfo := rest.ShareToShareInfo(fileShare)

	writeJson(w, http.StatusOK, shareInfo)
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
func setSharePublic(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r, true)
	if SafeErrorAndExit(err, w) {
		return
	}

	publicStr := r.URL.Query().Get("public")
	public := publicStr == "true"
	if !public && publicStr != "false" {
		writeJson(w, http.StatusBadRequest, rest.WeblensErrorInfo{Error: "Public must be true or false"})
		return
	}

	share, err := getShareFromCtx[*models.FileShare](w, r)
	if err != nil {
		return
	}

	if !pack.AccessService.CanUserModifyShare(u, share) {
		SafeErrorAndExit(werror.ErrNoShareAccess, w)
		return
	}

	err = pack.ShareService.SetSharePublic(share, public)
	if SafeErrorAndExit(err, w) {
		return
	}

	w.WriteHeader(http.StatusOK)
}

// SetShareAccessors godoc
//
//	@ID			SetShareAccessors
//
//	@Summary	Update a share's accessors list
//	@Tags		Share
//	@Produce	json
//	@Param		shareId	path	string				true	"Share Id"
//	@Param		request	body	rest.UserListBody	true	"Share Accessors"
//	@Success	200
//	@Failure	404
//	@Router		/share/{shareId}/accessors [patch]
func setShareAccessors(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r, true)
	if SafeErrorAndExit(err, w) {
		return
	}

	share, err := getShareFromCtx[*models.FileShare](w, r)
	if err != nil {
		return
	}

	if !pack.AccessService.CanUserModifyShare(u, share) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ub, err := readCtxBody[rest.UserListBody](w, r)
	if err != nil {
		log.ShowErr(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var addUsers []*models.User
	for _, un := range ub.AddUsers {
		u := pack.UserService.Get(un)
		if u == nil {
			writeJson(w, http.StatusNotFound, rest.WeblensErrorInfo{Error: "Could not find user with name " + un})
			return
		}
		addUsers = append(addUsers, u)
	}

	var removeUsers []*models.User
	for _, un := range ub.RemoveUsers {
		u := pack.UserService.Get(un)
		if u == nil {
			writeJson(w, http.StatusNotFound, rest.WeblensErrorInfo{Error: "Could not find user with name " + un})
			return
		}
		removeUsers = append(removeUsers, u)
	}

	if len(addUsers) > 0 {
		err = pack.ShareService.AddUsers(share, addUsers)
		if err != nil {
			log.ShowErr(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if len(removeUsers) > 0 {
		err = pack.ShareService.RemoveUsers(share, removeUsers)
		if err != nil {
			log.ShowErr(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
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
func deleteShare(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	shareId := models.ShareId(chi.URLParam(r, "shareId"))

	s := pack.ShareService.Get(shareId)
	if s == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err := pack.ShareService.Del(s.ID())
	if err != nil {
		log.ErrTrace(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
