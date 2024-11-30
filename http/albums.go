package http

import (
	"encoding/json"
	"net/http"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/rest"
	"github.com/go-chi/chi/v5"
)

// GetAlbum godoc
//
//	@ID			GetAlbum
//
//	@Security	SessionAuth
//
//	@Summary	Get album by album Id
//	@Tags		Album
//	@Produce	json
//	@Param		albumId	path		string			true	"Album Id"
//	@Param		shareId	query		string			false	"Share Id"
//	@Success	200		{object}	rest.AlbumInfo	"Album Info"
//	@Failure	404
//	@Router		/albums/{albumId} [get]
func getAlbum(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}
	if u == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	sh, err := getShareFromCtx[*models.AlbumShare](w, r)

	if SafeErrorAndExit(err, w) {
		return
	}

	albumId := models.AlbumId(chi.URLParam(r, "albumId"))
	album := pack.AlbumService.Get(albumId)
	if album == nil {
		SafeErrorAndExit(werror.ErrNoAlbum, w)
		return
	}

	if !pack.AccessService.CanUserAccessAlbum(u, album, sh) {
		SafeErrorAndExit(werror.ErrNoAlbumAccess, w)
		return
	}

	// raw := r.URL.Query().Get("raw") == "true"
	//
	// var medias []*models.Media
	// for media := range pack.AlbumService.GetAlbumMedias(album) {
	// 	if media == nil {
	// 		continue
	// 	}
	// 	if !raw && pack.MediaService.GetMediaType(media).IsRaw() {
	// 		continue
	// 	}
	// 	medias = append(medias, media)
	// }

	AlbumInfo := rest.AlbumToAlbumInfo(album)
	writeJson(w, http.StatusOK, AlbumInfo)
}

// GetAlbums godoc
//
//	@ID			GetAlbums
//
//	@Security	SessionAuth
//
//	@Summary	Get albums for a user
//	@Tags		Album
//	@Produce	json
//	@Param		filter	query	string			false	"Albums filter"
//	@Success	200		{array}	rest.AlbumInfo	"Album Infos"
//	@Failure	404
//	@Router		/albums [get]
func getAlbums(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}

	albums, err := pack.AlbumService.GetAllByUser(u)
	if SafeErrorAndExit(err, w) {
		return
	}

	// includeShared := r.URL.Query().Get("includeShared")

	filterString := r.URL.Query().Get("filter")
	var filter []string
	if filterString != "" {
		err := json.Unmarshal([]byte(filterString), &filter)
		if SafeErrorAndExit(err, w) {
			return
		}
	}

	var albumInfos []rest.AlbumInfo
	for _, a := range albums {
		albumInfos = append(albumInfos, rest.AlbumToAlbumInfo(a))
	}

	writeJson(w, http.StatusOK, albumInfos)
}

// CreateAlbum godoc
//
//	@ID			CreateAlbum
//
//	@Security	SessionAuth
//
//	@Summary	Create a new album
//	@Tags		Album
//	@Produce	json
//	@Param		NewAlbumParams	query		rest.CreateAlbumParams	true	"Create Album Params"
//	@Success	200				{object}	rest.AlbumInfo			"Album Info"
//	@Router		/albums [post]
func createAlbum(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}

	albumData, err := readCtxBody[rest.CreateAlbumParams](w, r)
	if err != nil {
		return
	}

	newAlbum := models.NewAlbum(albumData.Name, u)
	err = pack.AlbumService.Add(newAlbum)
	if err != nil {
		log.ShowErr(err)
		writeJson(w, http.StatusInternalServerError, rest.WeblensErrorInfo{Error: "Album creation failed"})
	}

	writeJson(w, http.StatusOK, newAlbum)
}

// UpdateAlbum godoc
//
//	@ID			UpdateAlbum
//
//	@Security	SessionAuth
//
//	@Summary	Update an Album
//	@Tags		Album
//	@Produce	json
//	@Param		albumId				path	string					true	"Album Id"
//	@Param		UpdateAlbumParams	query	rest.UpdateAlbumParams	true	"Update Album Params"
//	@Success	200
//	@Router		/albums/{albumId} [patch]
func updateAlbum(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}
	sh, err := getShareFromCtx[*models.AlbumShare](w, r)
	if SafeErrorAndExit(err, w) {
		return
	}

	albumId := models.AlbumId(chi.URLParam(r, "albumId"))
	a := pack.AlbumService.Get(albumId)
	if a == nil {
		writeJson(w, http.StatusNotFound, rest.WeblensErrorInfo{Error: "Album not found"})
		return
	}

	if a.GetOwner() != u.GetUsername() {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	update, err := readCtxBody[rest.UpdateAlbumParams](w, r)
	if err != nil {
		return
	}

	var ms []*models.Media
	if len(update.AddMedia) != 0 {
		ms = internal.FilterMap(
			update.AddMedia, func(mId models.ContentId) (*models.Media, bool) {
				m := pack.MediaService.Get(mId)
				if m != nil {
					return m, true
				} else {
					return m, false
				}
			},
		)

		if len(ms) == 0 {
			writeJson(w, http.StatusBadRequest, rest.WeblensErrorInfo{Error: "No valid media Ids in request"})
			return
		}
	}

	if len(update.AddFolders) != 0 {
		folders := internal.Map(
			update.AddFolders, func(fId fileTree.FileId) *fileTree.WeblensFileImpl {
				f, err := pack.FileService.GetFileSafe(fId, u, nil)
				if err != nil {
					log.ShowErr(err)
					return nil
				}
				return f
			},
		)

		ms = append(ms, pack.MediaService.RecursiveGetMedia(folders...)...)
	}

	// addedCount := 0
	if len(ms) != 0 {
		err = pack.AlbumService.AddMediaToAlbum(a, ms...)
		if err != nil {
			log.ErrTrace(err)
			writeJson(w, http.StatusInternalServerError, rest.WeblensErrorInfo{Error: "Failed to add media to album"})
			return
		}

		if a.GetCover() == "" {
			err = pack.AlbumService.SetAlbumCover(a.ID(), ms[0])
			if err != nil {
				log.ErrTrace(err)
				writeJson(w, http.StatusInternalServerError, rest.WeblensErrorInfo{Error: "Failed to set album cover"})
				return
			}
		}

	}

	if update.RemoveMedia != nil {
		err = pack.AlbumService.RemoveMediaFromAlbum(a, update.RemoveMedia...)
		if err != nil {
			log.ErrTrace(err)
			writeJson(w, http.StatusInternalServerError, rest.WeblensErrorInfo{Error: "Failed to remove media from album"})
			return
		}
	}

	if update.Cover != "" {
		cover := pack.MediaService.Get(update.Cover)
		if cover == nil {
			writeJson(w, http.StatusNotFound, rest.WeblensErrorInfo{Error: "Cover id not found"})
			return
		}
		err = pack.AlbumService.SetAlbumCover(a.ID(), cover)
		if err != nil {
			log.ErrTrace(err)
			writeJson(w, http.StatusInternalServerError, rest.WeblensErrorInfo{Error: "Failed to set album cover"})
			return
		}
	}

	if update.NewName != "" {
		err := pack.AlbumService.RenameAlbum(a, update.NewName)
		if err != nil {
			log.ErrTrace(err)
			writeJson(w, http.StatusInternalServerError, rest.WeblensErrorInfo{Error: "Failed to set album name"})
			return
		}
	}

	if len(update.RemoveUsers) != 0 {
		var users []*models.User
		for _, username := range update.RemoveUsers {
			users = append(users, pack.UserService.Get(username))
		}

		err = pack.ShareService.RemoveUsers(sh, users)

		if err != nil {
			log.ErrTrace(err)
			writeJson(w, http.StatusInternalServerError, rest.WeblensErrorInfo{Error: "Failed to un-share user(s)"})
			return
		}
	}

	if len(update.Users) != 0 {
		var users []*models.User
		for _, username := range update.RemoveUsers {
			users = append(users, pack.UserService.Get(username))
		}

		err = pack.ShareService.AddUsers(sh, users)
		if err != nil {
			log.ErrTrace(err)
			writeJson(w, http.StatusInternalServerError, rest.WeblensErrorInfo{Error: "Failed to share user(s)"})
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteOrLeaveAlbum godoc
//
//	@ID			DeleteOrLeaveAlbum
//
//	@Security	SessionAuth
//
//	@Summary	Delete an Album, or unshare it if the user is not the owner
//	@Tags		Album
//	@Produce	json
//	@Param		albumId	path	string	true	"Album Id"
//	@Param		shareId	query	string	false	"Share Id"
//	@Success	200
//	@Router		/albums/{albumId} [delete]
func deleteAlbum(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}
	sh, err := getShareFromCtx[*models.AlbumShare](w, r)
	if SafeErrorAndExit(err, w) {
		return
	}

	if u == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	albumId := models.AlbumId(chi.URLParam(r, "albumId"))

	a := pack.AlbumService.Get(albumId)

	// err or user does not have access to this album, claim not found
	if a == nil || !pack.AccessService.CanUserAccessAlbum(u, a, sh) {
		SafeErrorAndExit(werror.ErrNoAlbumAccess, w)
		return
	}

	// If the user is not the owner, then unshare them from the album
	if a.GetOwner() != u.GetUsername() {
		err = pack.ShareService.RemoveUsers(sh, []*models.User{u})
		if SafeErrorAndExit(err, w) {
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	err = pack.AlbumService.Del(albumId)
	if SafeErrorAndExit(err, w) {
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetAlbumMedia godoc
//
//	@ID	GetAlbumMedia
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Get media in an album
//	@Tags		Album
//	@Produce	json
//	@Param		albumId	path	string			true	"Album Id"
//	@Param		shareId	query	string			false	"Share Id"
//	@Success	200		{array}	rest.MediaInfo	"Media Info"
//	@Success	404
//	@Router		/albums/{albumId}/media [get]
func getAlbumMedia(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}
	sh, err := getShareFromCtx[*models.AlbumShare](w, r)
	if SafeErrorAndExit(err, w) {
		return
	}

	albumId := models.AlbumId(chi.URLParam(r, "albumId"))

	a := pack.AlbumService.Get(albumId)

	// err or user does not have access to this album, claim not found
	if a == nil || !pack.AccessService.CanUserAccessAlbum(u, a, sh) {
		SafeErrorAndExit(werror.ErrNoAlbumAccess, w)
		return
	}

	var mediaInfos []rest.MediaInfo
	for _, mId := range a.Medias {
		m := pack.MediaService.Get(mId)
		if m == nil {
			continue
		}
		mediaInfos = append(mediaInfos, rest.MediaToMediaInfo(m))
	}

	writeJson(w, http.StatusOK, mediaInfos)
}

// func unshareMeAlbum(w http.ResponseWriter, r *http.Request) {
// 	pack := getServices(r)
// 	u, err := getUserFromCtx(r)
// 	if SafeErrorAndExit(err, w) {
// 		return
// 	}
// 	sh, err := getShareFromCtx[*models.AlbumShare](w, r)
// 	if SafeErrorAndExit(err, w) {
// 		return
// 	}
//
// 	albumId := models.AlbumId(chi.URLParam(r, "albumId"))
// 	a := pack.AlbumService.Get(albumId)
// 	if a == nil {
// 		w.WriteHeader(http.StatusNotFound)
// 		return
// 	}
//
// 	if !pack.AccessService.CanUserAccessAlbum(u, a, sh) {
// 		w.WriteHeader(http.StatusNotFound)
// 		return
// 	}
//
// 	err = pack.ShareService.RemoveUsers(sh, []*models.User{u})
// 	if err != nil {
// 		log.ShowErr(err)
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}
//
// 	w.WriteHeader(http.StatusOK)
// }

// func albumPreviewMedia(w http.ResponseWriter, r *http.Request) {
// 	pack := getServices(r)
// 	albumId := models.AlbumId(chi.URLParam(r, "albumId"))
//
// 	a := pack.AlbumService.Get(albumId)
// 	if a == nil {
// 		w.WriteHeader(http.StatusNotFound)
// 		return
// 	}
//
// 	albumMs := slices.Collect(pack.AlbumService.GetAlbumMedias(a))
// 	randomMs := make([]models.ContentId, 0, 9)
//
// 	for len(albumMs) != 0 && len(randomMs) < 9 {
// 		index := rand.Intn(len(albumMs))
// 		m := pack.MediaService.Get(albumMs[index].ID())
// 		if m != nil && pack.MediaService.GetMediaType(m).IsRaw() && m.ID() != a.GetCover() {
// 			randomMs = append(randomMs, m.ID())
// 		}
//
// 		albumMs = internal.Banish(albumMs, index)
// 	}
//
// 	writeJson(w, http.StatusOK, gin.H{"mediaIds": randomMs})
// }
