package http

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"slices"

	"github.com/ethrousseau/weblens/api"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/fileTree"
	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/gin-gonic/gin"
	"github.com/modern-go/reflect2"
)

func getAlbum(ctx *gin.Context) {
	u := getUserFromCtx(ctx)
	if u == nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	
	albumData := AlbumService.Get(weblens.AlbumId(ctx.Param("weblens.AlbumId")))
	if albumData == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	if albumData.GetOwner() != u.GetUsername() && !slices.Contains(
		albumData.GetSharedWith(), u.GetUsername(),
	) {
		ctx.Status(http.StatusNotFound)
		return
	}

	raw := ctx.Query("raw") == "true"
	medias := albumData.GetMedias()
	medias = internal.Filter(
		medias, func(m *weblens.Media) bool {
			wlog.Debug.Println(m)
			if reflect2.IsNil(m) || m.GetMediaType() == nil {
				return false
			}
			return true
		},
	)
	if !raw {
		medias = internal.Filter(
			medias, func(m *weblens.Media) bool {
				wlog.Debug.Println(m)
				if reflect2.IsNil(m) || m.GetMediaType() == nil {
					return false
				}
				return !m.GetMediaType().IsRaw()
			},
		)
	}

	wlog.Debug.Println(medias)
	ctx.JSON(http.StatusOK, gin.H{"albumMeta": albumData, "medias": medias})
}

func getAlbums(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	if user == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	albums := AlbumService.GetAllByUser(user)
	includeShared := ctx.Query("includeShared")

	filterString := ctx.Query("filter")
	var filter []string
	if filterString != "" {
		err := json.Unmarshal([]byte(filterString), &filter)
		if err != nil {
			wlog.ShowErr(err)
			ctx.Status(http.StatusBadRequest)
			return
		}
	}
	var e bool
	albums = internal.Filter(
		albums, func(a *weblens.Album,;oooooooooooooooooooooooooooooooooooooooo ) bool {
			if includeShared == "false" && a.GetOwner() != user {
				return false
			}
			if len(filter) != 0 {
				filter, _, e = internal.YoinkFunc(
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

	newAlbum := weblens.NewAlbum(albumData.Name, user)
	err = AlbumService.Add(newAlbum)
	if err != nil {
		wlog.ShowErr(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Album creation failed"})
	}

	ctx.JSON(http.StatusOK, newAlbum)
}

func updateAlbum(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	weblens.AlbumId := weblens.AlbumId(ctx.Param("weblens.AlbumId"))
	a := AlbumService.Get(weblens.AlbumId)
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

	var ms []*weblens.Media
	if update.AddMedia != nil && len(update.AddMedia) != 0 {
		ms = internal.FilterMap(
			update.AddMedia, func(mId weblens.ContentId) (*weblens.Media, bool) {
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
		folders := internal.Map(
			update.AddFolders, func(fId types.FileId) *fileTree.WeblensFile {
				return types.SERV.FileTree.Get(fId)
			},
		)

		ms = append(
			ms, internal.Map(
				weblens.RecursiveGetMedia(types.SERV.MediaRepo, folders...),
				func(mId weblens.ContentId) *weblens.Media {
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
			wlog.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add media to album"})
			return
		}

		if a.GetCover() == "" {
			err = AlbumService.SetAlbumCover(a.ID(), ms[0])
			if err != nil {
				wlog.ErrTrace(err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set album cover"})
				return
			}
		}

	}

	if update.RemoveMedia != nil {
		err = a.RemoveMedia(update.RemoveMedia...)
		if err != nil {
			wlog.ErrTrace(err)
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
		err = AlbumService.SetAlbumCover(a.ID(), cover)
		if err != nil {
			wlog.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set album cover"})
			return
		}
	}

	if update.NewName != "" {
		err = a.Rename(update.NewName)
		if err != nil {
			wlog.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set album name"})
			return
		}
	}

	if len(update.RemoveUsers) != 0 {
		err = a.RemoveUsers(update.RemoveUsers...)
		if err != nil {
			wlog.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to un-share user(s)"})
			return
		}
	}

	if len(update.Users) != 0 {
		users := internal.Map(
			update.Users, func(u types.Username) types.User {
				return UserService.Get(u)
			},
		)

		err = a.AddUsers(users...)
		if err != nil {
			wlog.ErrTrace(err)
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

	weblens.AlbumId := weblens.AlbumId(ctx.Param("weblens.AlbumId"))

	a := AlbumService.Get(weblens.AlbumId)

	acc := dataStore.NewAccessMeta(user)
	// err or user does not have access to this album, claim not found
	if a == nil || !acc.CanAccessAlbum(a) {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
		return
	}

	// If the user is not the owner, then unshare them from the album
	if a.GetOwner() != user {
		err := a.RemoveUsers([]types.Username{user.GetUsername()}...)
		if err != nil {
			wlog.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to un-share user(s)"})
			return
		}
		ctx.Status(http.StatusOK)
		return
	}

	err := AlbumService.Del(weblens.AlbumId)
	if err != nil {
		wlog.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.Status(http.StatusOK)
}

func unshareMeAlbum(ctx *gin.Context) {
	user := getUserFromCtx(ctx)
	weblens.AlbumId := weblens.AlbumId(ctx.Param("weblens.AlbumId"))
	a := AlbumService.Get(weblens.AlbumId)
	if a == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	acc := dataStore.NewAccessMeta(user)
	if !acc.CanAccessAlbum(a) {
		ctx.Status(http.StatusNotFound)
		return
	}

	err := a.RemoveUsers(user.GetUsername())
	if err != nil {
		wlog.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

func albumPreviewMedia(ctx *gin.Context) {
	weblens.AlbumId := weblens.AlbumId(ctx.Param("weblens.AlbumId"))

	a := AlbumService.Get(weblens.AlbumId)
	if a == nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	albumMs := a.GetMedias()
	randomMs := make([]weblens.ContentId, 0, 9)

	for len(albumMs) != 0 && len(randomMs) < 9 {
		index := rand.Intn(len(albumMs))
		m := types.SERV.MediaRepo.Get(albumMs[index].ID())
		if m != nil && !m.GetMediaType().IsRaw() && m.ID() != a.GetCover() {
			randomMs = append(randomMs, m.ID())
		}

		albumMs = internal.Banish(albumMs, index)
	}

	ctx.JSON(http.StatusOK, gin.H{"mediaIds": randomMs})
}
