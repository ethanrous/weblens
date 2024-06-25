package routes

import (
	"encoding/json"
	"math/rand"
	"net/http"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/dataStore/album"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

func getAlbum(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	albumData := types.SERV.AlbumManager.Get(types.AlbumId(ctx.Param("albumId")))
	if albumData == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	raw := ctx.Query("raw") == "true"
	medias := albumData.GetMedias()
	if !raw {
		medias = util.Filter(
			medias, func(t types.Media) bool {
				return !t.GetMediaType().IsRaw()
			},
		)
	}

	ctx.JSON(http.StatusOK, gin.H{"albumMeta": albumData, "medias": medias})
}

func getAlbums(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	albums := types.SERV.AlbumManager.GetAllByUser(user)
	includeShared := ctx.Query("includeShared")

	filterString := ctx.Query("filter")
	var filter []string
	if filterString != "" {
		err := json.Unmarshal([]byte(filterString), &filter)
		if err != nil {
			util.ShowErr(err)
			ctx.Status(http.StatusBadRequest)
			return
		}
	}
	var e bool
	albums = util.Filter(
		albums, func(a types.Album) bool {
			if includeShared == "false" && a.GetOwner() != user {
				return false
			}
			if len(filter) != 0 {
				filter, _, e = util.YoinkFunc(
					filter, func(s string) bool {
						return s == a.GetName()
					},
				)
				return e
			}
			return true
		},
	)

	ctx.JSON(http.StatusOK, gin.H{"albums": albums})
}

func createAlbum(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	albumData, err := readCtxBody[albumCreateBody](ctx)
	if err != nil {
		return
	}

	newAlbum := album.New(albumData.Name, user)
	err = types.SERV.AlbumManager.Add(newAlbum)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Album creation failed"})
	}

	ctx.JSON(http.StatusOK, newAlbum)
}

func updateAlbum(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	albumId := types.AlbumId(ctx.Param("albumId"))
	a := types.SERV.AlbumManager.Get(albumId)
	if a == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
		return
	}

	if a.GetOwner() != user {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	update, err := readCtxBody[updateAlbumBody](ctx)
	if err != nil {
		return
	}
	
	var ms []types.Media
	if update.AddMedia != nil && len(update.AddMedia) != 0 {
		ms = util.FilterMap(
			update.AddMedia, func(mId types.ContentId) (types.Media, bool) {
				m := types.SERV.MediaRepo.Get(mId)
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
		folders := util.Map(
			update.AddFolders, func(fId types.FileId) types.WeblensFile {
				return types.SERV.FileTree.Get(fId)
			},
		)

		ms = append(
			ms, util.Map(
				dataStore.RecursiveGetMedia(types.SERV.MediaRepo, folders...), func(mId types.ContentId) types.Media {
					m := types.SERV.MediaRepo.Get(mId)
					return m
				},
			)...,
		)
	}

	addedCount := 0
	if len(ms) != 0 {
		err = a.AddMedia(ms...)
		if err != nil {
			util.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add media to album"})
			return
		}

		if a.GetCover() == nil || a.GetPrimaryColor() == "" {
			if a.GetCover() == nil {
				err = a.SetCover(ms[0])
				if err != nil {
					util.ErrTrace(err)
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set album cover"})
					return
				}
			}
		}
	}

	if update.RemoveMedia != nil {
		err = a.RemoveMedia(update.RemoveMedia...)
		if err != nil {
			util.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove media from album"})
			return
		}
	}

	if update.Cover != "" {
		cover := types.SERV.MediaRepo.Get(update.Cover)
		if cover == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Cover id not found"})
			return
		}
		err = a.SetCover(cover)
		if err != nil {
			util.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set album cover"})
			return
		}
	}

	if update.NewName != "" {
		err = a.Rename(update.NewName)
		if err != nil {
			util.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set album name"})
			return
		}
	}

	if len(update.RemoveUsers) != 0 {
		err = a.RemoveUsers(update.RemoveUsers...)
		if err != nil {
			util.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to un-share user(s)"})
			return
		}
	}

	if len(update.Users) != 0 {
		users := util.Map(
			update.Users, func(u types.Username) types.User {
				return types.SERV.UserService.Get(u)
			},
		)

		err = a.AddUsers(users...)
		if err != nil {
			util.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share user(s)"})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"errors": []string{}, "addedCount": addedCount})
}

func deleteAlbum(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	albumId := types.AlbumId(ctx.Param("albumId"))

	a := types.SERV.AlbumManager.Get(albumId)

	acc := dataStore.NewAccessMeta(user, types.SERV.FileTree)
	// err or user does not have access to this album, claim not found
	if a == nil || !acc.CanAccessAlbum(a) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
		return
	}

	// If the user is not the owner, then unshare them from the album
	if a.GetOwner() != user {
		err := a.RemoveUsers([]types.Username{user.GetUsername()}...)
		if err != nil {
			util.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to un-share user(s)"})
			return
		}
		ctx.Status(http.StatusOK)
		return
	}

	err := types.SERV.AlbumManager.Del(albumId)
	if err != nil {
		util.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.Status(http.StatusOK)
}

func unshareMeAlbum(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	albumId := types.AlbumId(ctx.Param("albumId"))
	a := types.SERV.AlbumManager.Get(albumId)
	if a == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	acc := dataStore.NewAccessMeta(user, types.SERV.FileTree)
	if !acc.CanAccessAlbum(a) {
		ctx.Status(http.StatusNotFound)
		return
	}

	err := a.RemoveUsers(user.GetUsername())
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func albumPreviewMedia(ctx *gin.Context) {
	albumId := types.AlbumId(ctx.Param("albumId"))

	a := types.SERV.AlbumManager.Get(albumId)
	if a == nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	albumMs := a.GetMedias()
	randomMs := make([]types.ContentId, 0, 9)

	for len(albumMs) != 0 && len(randomMs) < 9 {
		index := rand.Intn(len(albumMs))
		m := types.SERV.MediaRepo.Get(albumMs[index].ID())
		if m != nil && !m.GetMediaType().IsRaw() && m != a.GetCover() {
			randomMs = append(randomMs, m.ID())
		}

		albumMs = util.Banish(albumMs, index)
	}

	ctx.JSON(http.StatusOK, gin.H{"mediaIds": randomMs})
}
