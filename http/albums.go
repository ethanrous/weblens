package http

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"slices"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/gin-gonic/gin"
)

func getAlbum(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if u == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	sh, err := getShareFromCtx[*models.AlbumShare](ctx)

	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	album := pack.AlbumService.Get(models.AlbumId(ctx.Param("albumId")))
	if album == nil {
		ctx.JSON(http.StatusNotFound, werror.ErrNoAlbum)
		return
	}

	if !pack.AccessService.CanUserAccessAlbum(u, album, sh) {
		ctx.JSON(http.StatusNotFound, werror.ErrNoAlbum)
		return
	}

	raw := ctx.Query("raw") == "true"

	var medias []*models.Media
	for media := range pack.AlbumService.GetAlbumMedias(album) {
		if media == nil {
			continue
		}
		if !raw && pack.MediaService.GetMediaType(media).IsRaw() {
			continue
		}
		medias = append(medias, media)
	}

	ctx.JSON(http.StatusOK, gin.H{"albumMeta": album, "medias": medias})
}

func getAlbums(ctx *gin.Context) {
	pack := getServices(ctx)
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	albums, err := pack.AlbumService.GetAllByUser(user)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	// includeShared := ctx.Query("includeShared")

	filterString := ctx.Query("filter")
	var filter []string
	if filterString != "" {
		err := json.Unmarshal([]byte(filterString), &filter)
		if err != nil {
			log.ShowErr(err)
			ctx.Status(http.StatusBadRequest)
			return
		}
	}
	// var e bool
	// albums = internal.Filter(
	// 	albums, func(a *models.Album) bool {
	// 		if includeShared == "false" && a.GetOwner() != user.GetUsername() {
	// 			return false
	// 		}
	// 		if len(filter) != 0 {
	// 			filter, _, e = internal.YoinkFunc(
	// 				filter, func(s string) bool {
	// 					return s == a.GetName()
	// 				},
	// 			)
	// 			return e
	// 		}
	// 		return true
	// 	},
	// )

	ctx.JSON(http.StatusOK, gin.H{"albums": albums})
}

func createAlbum(ctx *gin.Context) {
	pack := getServices(ctx)
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	albumData, err := readCtxBody[albumCreateBody](ctx)
	if err != nil {
		return
	}

	newAlbum := models.NewAlbum(albumData.Name, user)
	err = pack.AlbumService.Add(newAlbum)
	if err != nil {
		log.ShowErr(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Album creation failed"})
	}

	ctx.JSON(http.StatusOK, newAlbum)
}

func updateAlbum(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	sh, err := getShareFromCtx[*models.AlbumShare](ctx)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	albumId := models.AlbumId(ctx.Param("albumId"))
	a := pack.AlbumService.Get(albumId)
	if a == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
		return
	}

	if a.GetOwner() != u.GetUsername() {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	update, err := readCtxBody[updateAlbumBody](ctx)
	if err != nil {
		return
	}

	var ms []*models.Media
	if update.AddMedia != nil && len(update.AddMedia) != 0 {
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
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "No valid media Ids in request"})
			return
		}
	}

	if update.AddFolders != nil && len(update.AddFolders) != 0 {
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

		ms = append(
			ms, internal.Map(
				pack.MediaService.RecursiveGetMedia(folders...),
				func(mId models.ContentId) *models.Media {
					m := pack.MediaService.Get(mId)
					return m
				},
			)...,
		)
	}

	addedCount := 0
	if len(ms) != 0 {
		err = pack.AlbumService.AddMediaToAlbum(a, ms...)
		if err != nil {
			log.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add media to album"})
			return
		}

		if a.GetCover() == "" {
			err = pack.AlbumService.SetAlbumCover(a.ID(), ms[0])
			if err != nil {
				log.ErrTrace(err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set album cover"})
				return
			}
		}

	}

	if update.RemoveMedia != nil {
		err = pack.AlbumService.RemoveMediaFromAlbum(a, update.RemoveMedia...)
		if err != nil {
			log.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove media from album"})
			return
		}
	}

	if update.Cover != "" {
		cover := pack.MediaService.Get(update.Cover)
		if cover == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Cover id not found"})
			return
		}
		err = pack.AlbumService.SetAlbumCover(a.ID(), cover)
		if err != nil {
			log.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set album cover"})
			return
		}
	}

	if update.NewName != "" {
		err := pack.AlbumService.RenameAlbum(a, update.NewName)
		if err != nil {
			log.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set album name"})
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
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to un-share user(s)"})
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
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share user(s)"})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"errors": []string{}, "addedCount": addedCount})
}

func deleteAlbum(ctx *gin.Context) {
	pack := getServices(ctx)
	user := getUserFromCtx(ctx)
	sh, err := getShareFromCtx[*models.AlbumShare](ctx)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	if user == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	albumId := models.AlbumId(ctx.Param("albumId"))

	a := pack.AlbumService.Get(albumId)

	// err or user does not have access to this album, claim not found
	if a == nil || !pack.AccessService.CanUserAccessAlbum(user, a, sh) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
		return
	}

	// If the user is not the owner, then unshare them from the album
	if a.GetOwner() != user.GetUsername() {
		err = pack.ShareService.RemoveUsers(sh, []*models.User{user})
		if err != nil {
			log.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to un-share user(s)"})
			return
		}
		ctx.Status(http.StatusOK)
		return
	}

	err = pack.AlbumService.Del(albumId)
	if err != nil {
		log.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.Status(http.StatusOK)
}

func unshareMeAlbum(ctx *gin.Context) {
	pack := getServices(ctx)
	user := getUserFromCtx(ctx)
	sh, err := getShareFromCtx[*models.AlbumShare](ctx)
	if err != nil {
		safe, code := werror.TrySafeErr(err)
		ctx.JSON(code, safe)
		return
	}

	albumId := models.AlbumId(ctx.Param("albumId"))
	a := pack.AlbumService.Get(albumId)
	if a == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	if !pack.AccessService.CanUserAccessAlbum(user, a, sh) {
		ctx.Status(http.StatusNotFound)
		return
	}

	err = pack.ShareService.RemoveUsers(sh, []*models.User{user})
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func albumPreviewMedia(ctx *gin.Context) {
	pack := getServices(ctx)
	albumId := models.AlbumId(ctx.Param("albumId"))

	a := pack.AlbumService.Get(albumId)
	if a == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	albumMs := slices.Collect(pack.AlbumService.GetAlbumMedias(a))
	randomMs := make([]models.ContentId, 0, 9)

	for len(albumMs) != 0 && len(randomMs) < 9 {
		index := rand.Intn(len(albumMs))
		m := pack.MediaService.Get(albumMs[index].ID())
		if m != nil && pack.MediaService.GetMediaType(m).IsRaw() && m.ID() != a.GetCover() {
			randomMs = append(randomMs, m.ID())
		}

		albumMs = internal.Banish(albumMs, index)
	}

	ctx.JSON(http.StatusOK, gin.H{"mediaIds": randomMs})
}
