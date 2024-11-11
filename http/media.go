package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ethanrous/bimg"
	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/rest"
	"github.com/gin-gonic/gin"
)

// GetMedia godoc
//
//	@ID			GetMedia
//
//	@Summary	Get paginated media
//	@Tags		Media
//	@Produce	json
//	@Param		raw			query		bool				false	"Include raw files"						Enums(true, false)	default(false)
//	@Param		hidden		query		bool				false	"Include hidden media"					Enums(true, false)	default(false)
//	@Param		sort		query		string				false	"Sort by field"							Enums(createDate)	default(createDate)
//	@Param		folderIds	query		string				false	"Search only in given folders"			SchemaExample([fId1, fId2])
//	@Param		mediaIds	query		string				false	"Get only media with the provided ids"	SchemaExample([mId1, id2])
//	@Success	200			{object}	rest.MediaBatchInfo	"Media Batch"
//	@Success	400
//	@Success	500
//	@Router		/media [get]
func getMediaBatch(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)

	folderIdsStr := ctx.Query("folderIds")
	if folderIdsStr != "" {
		var folderIds []fileTree.FileId
		err := json.Unmarshal([]byte(folderIdsStr), &folderIds)
		if werror.SafeErrorAndExit(err, ctx) {
			return
		}

		getMediaInFolders(pack, u, folderIds, ctx)
		return
	}

	mediaIdsStr := ctx.Query("mediaIds")
	if mediaIdsStr != "" {
		var mediaIds []fileTree.FileId
		err := json.Unmarshal([]byte(mediaIdsStr), &mediaIds)
		if werror.SafeErrorAndExit(err, ctx) {
			return
		}

		var medias []*models.Media
		for _, mId := range mediaIds {
			medias = append(medias, pack.MediaService.Get(mId))
		}

		getMediaInFolders(pack, u, mediaIds, ctx)
		ctx.JSON(http.StatusOK, rest.MediaBatchInfo{Media: medias})
		return
	}

	sort := ctx.Query("sort")
	if sort == "" {
		sort = "createDate"
	}

	raw := ctx.Query("raw") == "true"
	hidden := ctx.Query("hidden") == "true"

	var page int64
	var err error
	pageStr := ctx.Query("page")
	if pageStr != "" {
		page, err = strconv.ParseInt(pageStr, 10, 32)
		if werror.SafeErrorAndExit(err, ctx) {
			return
		}
	} else {
		page = 0
	}

	var limit int64
	limitStr := ctx.Query("limit")
	if limitStr != "" {
		limit, err = strconv.ParseInt(limitStr, 10, 32)
		if werror.SafeErrorAndExit(err, ctx) {
			return
		}
	} else {
		limit = 100
	}

	var albumFilter []models.AlbumId
	albumsStr := ctx.Query("albums")
	if albumsStr != "" {
		err = json.Unmarshal([]byte(albumsStr), &albumFilter)
		if werror.SafeErrorAndExit(err, ctx) {
			return
		}
	}

	var mediaFilter []models.ContentId
	for _, albumId := range albumFilter {
		a := pack.AlbumService.Get(albumId)
		mediaFilter = append(mediaFilter, a.GetMedias()...)
	}

	ms, err := pack.MediaService.GetFilteredMedia(u, sort, 1, mediaFilter, raw, hidden)
	if werror.SafeErrorAndExit(err, ctx) {
		return
	}

	var slicedMs []*models.Media
	if (page+1)*limit > int64(len(ms)) {
		slicedMs = ms[(page)*limit:]
	} else {
		slicedMs = ms[(page)*limit : (page+1)*limit]
	}

	res := rest.MediaBatchInfo{Media: slicedMs, MediaCount: len(ms)}

	ctx.JSON(http.StatusOK, res)
}

// GetMediaTypes godoc
//
//	@ID			GetMediaTypes
//
//	@Summary	Get media type dictionary
//	@Tags		Media
//	@Produce	json
//	@Success	200	{object}	rest.MediaTypeInfo	"Media types"
//	@Router		/media/types  [get]
func getMediaTypes(ctx *gin.Context) {
	pack := getServices(ctx)
	ctx.JSON(http.StatusOK, pack.MediaService.GetMediaTypes())
}

// GetMediaInfo godoc
//
//	@ID			GetMediaInfo
//
//	@Summary	Get media info
//	@Tags		Media
//	@Produce	json
//	@Param		mediaId	path		string			true	"Media Id"
//	@Success	200		{object}	models.Media	"Media Info"
//	@Router		/media/{mediaId}/info [get]
func getMediaInfo(ctx *gin.Context) {
	pack := getServices(ctx)
	mediaId := ctx.Param("mediaId")
	m := pack.MediaService.Get(mediaId)
	if m == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, m)
}

// GetMediaImage godoc
//
//	@Id			GetMediaImage
//
//	@Summary	Get a media image bytes
//	@Tags		Media
//	@Produce	image/webp, image/png, image/jpeg
//	@Param		mediaId	path		string	true	"Media Id"
//	@Param		quality	query		string	true	"Image Quality"	Enums(thumbnail, fullres)
//	@Param		page	query		int		false	"Page number"
//	@Success	200		{string}	binary	"image bytes"
//	@Success	500
//	@Router		/media/{mediaId} [get]
func getMediaImage(ctx *gin.Context) {
	quality := models.MediaQuality(ctx.Query("quality"))
	getProcessedMedia(ctx, quality, "")
}

func streamVideo(ctx *gin.Context) {
	pack := getServices(ctx)
	// u := getUserFromCtx(ctx)
	sh, err := getShareFromCtx[*models.FileShare](ctx)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	mediaId := ctx.Param("mediaId")
	m := pack.MediaService.Get(mediaId)
	if m == nil {
		ctx.Status(http.StatusNotFound)
		return
	} else if !pack.MediaService.GetMediaType(m).IsVideo() {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "media is not of type video"})
		return
	}

	streamer, err := pack.MediaService.StreamVideo(m, pack.UserService.GetRootUser(), sh)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	chunkName := ctx.Param("chunkName")
	if chunkName != "" {
		ctx.File(streamer.GetEncodeDir() + chunkName)
		return
	}

	playlistFilePath := filepath.Join(streamer.GetEncodeDir(), "list.m3u8")
	for {
		_, err := os.Stat(playlistFilePath)
		if streamer.Err() != nil {
			log.ShowErr(streamer.Err())
			ctx.Status(http.StatusInternalServerError)
			return
		}
		if err != nil {
			time.Sleep(time.Millisecond * 100)
		} else {
			break
		}
	}
	ctx.File(playlistFilePath)
}

// SetMediaVisibility godoc
//
//	@Id			SetMediaVisibility
//
//	@Summary	Set media visibility
//	@Tags		Media
//	@Produce	json
//	@Param		hidden		query	bool				true	"Set the media visibility"	Enums(true, false)
//	@Param		mediaIds	body	rest.MediaIdsParams	true	"MediaIds to change visibility of"
//	@Success	200
//	@Success	404
//	@Success	500
//	@Router		/media/visibility [patch]
func hideMedia(ctx *gin.Context) {
	pack := getServices(ctx)
	body, err := readCtxBody[rest.MediaIdsParams](ctx)
	if err != nil {
		return
	}

	hidden := ctx.Query("hidden") == "true"

	medias := make([]*models.Media, len(body.MediaIds))
	for i, mId := range body.MediaIds {
		m := pack.MediaService.Get(mId)
		if m == nil {
			ctx.Status(http.StatusNotFound)
			return
		}
		medias[i] = m
	}

	for _, m := range medias {
		err = pack.MediaService.HideMedia(m, hidden)
		if err != nil {
			log.ShowErr(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}
	}

	ctx.Status(http.StatusOK)
}

func adjustMediaDate(ctx *gin.Context) {
	pack := getServices(ctx)
	body, err := readCtxBody[rest.MediaTimeBody](ctx)
	if err != nil {
		return
	}

	anchor := pack.MediaService.Get(body.AnchorId)
	if anchor == nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	extras := internal.Map(
		body.MediaIds, func(mId models.ContentId) *models.Media { return pack.MediaService.Get(body.AnchorId) },
	)

	err = pack.MediaService.AdjustMediaDates(anchor, body.NewTime, extras)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	ctx.Status(http.StatusOK)
}

func likeMedia(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if u == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	mediaId := ctx.Param("mediaId")
	liked := ctx.Query("liked") == "true"

	err := pack.MediaService.SetMediaLiked(mediaId, liked, u.GetUsername())
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}

// Helper function
func getMediaInFolders(pack *models.ServicePack, u *models.User, folderIds []string, ctx *gin.Context) {
	var folders []*fileTree.WeblensFileImpl
	for _, folderId := range folderIds {
		f, err := pack.FileService.GetFileSafe(folderId, u, nil)
		if err != nil {
			log.ShowErr(err)
			ctx.Status(http.StatusNotFound)
			return
		}
		folders = append(folders, f)
	}

	res := rest.MediaBatchInfo{Media: pack.MediaService.RecursiveGetMedia(folders...)}

	ctx.JSON(http.StatusOK, res)
}

// Helper function
func getProcessedMedia(ctx *gin.Context, q models.MediaQuality, format string) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	mediaId := ctx.Param("mediaId")

	m := pack.MediaService.Get(mediaId)
	if m == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Media with given ID not found"})
		return
	}

	var pageNum int
	var err error
	if q == models.HighRes && m.PageCount > 1 {
		pageString := ctx.Query("page")
		pageNum, err = strconv.Atoi(pageString)
		if err != nil {
			log.Debug.Println("Bad page number trying to get fullres multi-page image")
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "bad page number"})
			return
		}
	}

	if q == models.Video && !pack.MediaService.GetMediaType(m).IsVideo() {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "media type is not video"})
		return
	}

	bs, err := pack.MediaService.FetchCacheImg(m, q, pageNum)

	if errors.Is(err, werror.ErrNoCache) {
		f, err := pack.FileService.GetFileSafe(m.GetFiles()[0], u, nil)
		if err == nil {
			meta := models.ScanMeta{
				File:         f.GetParent(),
				FileService:  pack.FileService,
				MediaService: pack.MediaService,
				TaskService:  pack.TaskService,
				TaskSubber:   pack.ClientService,
				Caster:       pack.Caster,
			}
			_, err = pack.TaskService.DispatchJob(models.ScanDirectoryTask, meta, nil)
			if err != nil {
				log.ShowErr(err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to launch process media task"})
				return
			}
			ctx.Status(http.StatusNoContent)
			return
		} else {
			log.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get media content"})
			return
		}
	}

	if err != nil {
		log.ErrTrace(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get media content"})
		return
	}

	if format == "png" {
		image := bimg.NewImage(bs)
		bs, err = image.Convert(bimg.PNG)
		if err != nil {
			log.ErrTrace(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert image to PNG"})
			return
		}
	}

	ctx.Header("Cache-Control", "max-age=3600")

	ctx.Status(http.StatusOK)
	_, err = ctx.Writer.Write(bs)

	if err != nil {
		log.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
}
