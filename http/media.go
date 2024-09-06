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
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/gin-gonic/gin"
)

func getMediaBatch(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if u == nil {
		ctx.Status(http.StatusUnauthorized)
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
	if ctx.Query("page") != "" {
		page, err = strconv.ParseInt(ctx.Query("page"), 10, 32)
		if err != nil {
			ctx.Status(http.StatusBadRequest)
			return
		}
	} else {
		page = 0
	}

	var limit int64
	if ctx.Query("limit") != "" {
		limit, err = strconv.ParseInt(ctx.Query("limit"), 10, 32)
		if err != nil {
			ctx.Status(http.StatusBadRequest)
			return
		}
	} else {
		limit = 100
	}

	var albumFilter []models.AlbumId
	err = json.Unmarshal([]byte(ctx.Query("albums")), &albumFilter)
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	var mediaFilter []models.ContentId
	for _, albumId := range albumFilter {
		a := pack.AlbumService.Get(albumId)
		mediaFilter = append(mediaFilter, a.GetMedias()...)
	}

	ms, err := pack.MediaService.GetFilteredMedia(u, sort, 1, mediaFilter, raw, hidden)
	if err != nil {
		log.ErrTrace(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve media"})
		return
	}

	var slicedMs []*models.Media
	if (page+1)*limit > int64(len(ms)) {
		slicedMs = ms[(page)*limit:]
	} else {
		slicedMs = ms[(page)*limit : (page+1)*limit]
	}

	ctx.JSON(http.StatusOK, gin.H{"Media": slicedMs, "mediaCount": len(ms)})
}

func getMediaTypes(ctx *gin.Context) {
	pack := getServices(ctx)
	ctx.JSON(http.StatusOK, pack.MediaService.GetMediaTypes())
}

func getMediaInfo(ctx *gin.Context) {
	pack := getServices(ctx)
	mediaId := models.ContentId(ctx.Param("mediaId"))
	m := pack.MediaService.Get(mediaId)
	if m == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, m)
}

func getMediaThumbnail(ctx *gin.Context) {
	getProcessedMedia(ctx, models.LowRes, "")
}

func getMediaThumbnailPng(ctx *gin.Context) {
	getProcessedMedia(ctx, models.LowRes, "png")
}

func getMediaFullres(ctx *gin.Context) {
	getProcessedMedia(ctx, models.HighRes, "")
}

func streamVideo(ctx *gin.Context) {
	pack := getServices(ctx)
	// u := getUserFromCtx(ctx)
	sh, err := getShareFromCtx[*models.FileShare](ctx)
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	mediaId := models.ContentId(ctx.Param("mediaId"))
	m := pack.MediaService.Get(mediaId)
	if m == nil {
		ctx.Status(http.StatusNotFound)
		return
	} else if !pack.MediaService.GetMediaType(m).IsVideo() {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "media is not of type video"})
		return
	}

	// TODO - figure out how to send auth headers when getting video from client
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

func fetchMediaBytes(ctx *gin.Context) {
	q := models.MediaQuality(ctx.Query("quality"))
	getProcessedMedia(ctx, q, "")
}

func getProcessedMedia(ctx *gin.Context, q models.MediaQuality, format string) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	mediaId := models.ContentId(ctx.Param("mediaId"))

	var pageNum int
	var err error
	pageString := ctx.Query("page")
	if pageString != "" {
		pageNum, err = strconv.Atoi(pageString)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "bad page number"})
			return
		}
	}

	m := pack.MediaService.Get(mediaId)
	if m == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Media with given ID not found"})
		return
	}

	// if m.Owner != u.GetUsername() {
	// 	ctx.Status(http.StatusNotFound)
	// 	return
	// }

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

	ctx.Status(http.StatusOK)
	_, err = ctx.Writer.Write(bs)

	if err != nil {
		log.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
}

func hideMedia(ctx *gin.Context) {
	pack := getServices(ctx)
	body, err := readCtxBody[mediaIdsBody](ctx)
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
	body, err := readCtxBody[mediaTimeBody](ctx)
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

func getMediaArchive(ctx *gin.Context) {
	pack := getServices(ctx)
	ms := pack.MediaService.GetAll()
	ctx.JSON(http.StatusOK, ms)
}

func likeMedia(ctx *gin.Context) {
	pack := getServices(ctx)
	u := getUserFromCtx(ctx)
	if u == nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	mediaId := models.ContentId(ctx.Param("mediaId"))
	liked := ctx.Query("liked") == "true"

	err := pack.MediaService.SetMediaLiked(mediaId, liked, u.GetUsername())
	if err != nil {
		log.ShowErr(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusOK)
}
