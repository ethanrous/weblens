package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/dataStore/media"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

func getMediaBatch(ctx *gin.Context) {
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

	var albumFilter []types.AlbumId
	err = json.Unmarshal([]byte(ctx.Query("albums")), &albumFilter)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	ms, err := types.SERV.MediaRepo.GetFilteredMedia(u, sort, 1, albumFilter, raw, hidden)
	if err != nil {
		util.ErrTrace(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve media"})
		return
	}

	var slicedMs []types.Media
	if (page+1)*limit > int64(len(ms)) {
		slicedMs = ms[(page)*limit:]
	} else {
		slicedMs = ms[(page)*limit : (page+1)*limit]
	}

	ctx.JSON(http.StatusOK, gin.H{"Media": slicedMs, "mediaCount": len(ms)})
}

func getMediaTypes(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, types.SERV.MediaRepo.TypeService())
}

func getMediaInfo(ctx *gin.Context) {
	mediaId := types.ContentId(ctx.Param("mediaId"))
	m := types.SERV.MediaRepo.Get(mediaId)
	if m == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, m)
}

func getMediaThumbnail(ctx *gin.Context) {
	getProcessedMedia(ctx, types.Thumbnail)
}

func getMediaFullres(ctx *gin.Context) {
	getProcessedMedia(ctx, types.Fullres)
}

func streamVideo(ctx *gin.Context) {
	mediaId := types.ContentId(ctx.Param("mediaId"))
	m := types.SERV.MediaRepo.Get(mediaId)
	if m == nil {
		ctx.Status(http.StatusNotFound)
		return
	} else if !m.GetMediaType().IsVideo() {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "media is not of type video"})
		return
	}

	streamer := media.NewVideoStreamer(m).Encode()

	chunkName := ctx.Param("chunkName")
	if chunkName != "" {
		ctx.File(streamer.GetEncodeDir() + chunkName)
		return
	}

	playlistFilePath := filepath.Join(streamer.GetEncodeDir(), "list.m3u8")
	for {
		_, err := os.Stat(playlistFilePath)
		if streamer.Err() != nil {
			util.ShowErr(streamer.Err())
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
	// ctx.Header("Content-Type", "application/x-mpegURL")
	// fp, err := os.Open(playlistFilePath)
	// if err != nil {
	// 	util.ShowErr(err)
	// 	return
	// }
	// defer fp.Close()
	// _, err = io.Copy(ctx.Writer, fp)
	// if err != nil {
	// 	util.ShowErr(err)
	// }
}

func fetchMediaBytes(ctx *gin.Context) {
	q := types.Quality(ctx.Query("quality"))
	getProcessedMedia(ctx, q)
}

func getProcessedMedia(ctx *gin.Context, q types.Quality) {
	mediaId := types.ContentId(ctx.Param("mediaId"))

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

	m := types.SERV.MediaRepo.Get(mediaId)
	if m == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Media with given ID not found"})
		return
	}

	if q == types.Video && !m.GetMediaType().IsVideo() {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "media type is not video"})
		return
	}

	bs, err := m.ReadDisplayable(q, pageNum)

	if errors.Is(err, dataStore.ErrNoCache()) {
		f := types.SERV.FileTree.Get(m.GetFiles()[0])
		if f != nil {
			types.SERV.TaskDispatcher.ScanDirectory(f.GetParent(), types.SERV.Caster)
			ctx.Status(http.StatusNoContent)
			return
		} else {
			util.ErrTrace(types.WeblensErrorMsg("failed to get media content"))
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get media content"})
			return
		}
	}

	if err != nil {
		util.ErrTrace(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get media content"})
		return
	}

	_, err = ctx.Writer.Write(bs)

	if err != nil {
		util.ErrTrace(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
}

func hideMedia(ctx *gin.Context) {
	body, err := readCtxBody[mediaIdsBody](ctx)
	if err != nil {
		return
	}

	hidden := ctx.Query("hidden") == "true"

	medias := make([]types.Media, len(body.MediaIds))
	for i, mId := range body.MediaIds {
		m := types.SERV.MediaRepo.Get(mId)
		if m == nil {
			ctx.Status(http.StatusNotFound)
			return
		}
		medias[i] = m
	}

	for _, m := range medias {
		err := m.Hide(hidden)
		if err != nil {
			util.ShowErr(err)
			ctx.Status(http.StatusInternalServerError)
			return
		}
	}

	ctx.Status(http.StatusOK)
}

func adjustMediaDate(ctx *gin.Context) {
	body, err := readCtxBody[mediaTimeBody](ctx)
	if err != nil {
		return
	}

	anchor := types.SERV.MediaRepo.Get(body.AnchorId)
	if anchor == nil {
		ctx.Status(http.StatusNotFound)
		return
	}
	extras := util.Map(
		body.MediaIds, func(mId types.ContentId) types.Media { return types.SERV.MediaRepo.Get(body.AnchorId) },
	)

	err = media.AdjustMediaDates(anchor, body.NewTime, extras)
	if err != nil {
		util.ShowErr(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	ctx.Status(http.StatusOK)
}

func getMediaArchive(ctx *gin.Context) {
	ms := types.SERV.MediaRepo.GetAll()
	ctx.JSON(http.StatusOK, ms)
}
