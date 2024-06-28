package routes

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/dataStore/media"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/gin-gonic/gin"
)

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

	if true {
		requestRange := ctx.GetHeader("Range")
		startByte := 0
		endByte := 0

		if requestRange != "" {
			rangeParts := strings.Split(requestRange[6:], "-")
			startByte, _ = strconv.Atoi(rangeParts[0])
			endByte, _ = strconv.Atoi(rangeParts[1])
		}

		streamer := media.NewVideoStreamer(m)
		bufSize := streamer.PreLoadBuf()
		defer streamer.RelinquishStream()

		var rangeTotalStr = ""
		if bufSize == -1 {
			// ROUGH prediction of video size
			// predict := (128000 + util.GetVideoConstBitrate()) * m.GetVideoLength() / 8000
			// rangeTotalStr = strconv.Itoa(predict)
			rangeTotalStr = "*"
			for streamer.Len() == 0 {
				time.Sleep(time.Millisecond)
			}
			if endByte == 0 {
				endByte = streamer.Len() - 1
			}
		} else {
			rangeTotalStr = strconv.Itoa(bufSize)
			if endByte == 0 {
				endByte = bufSize - 1
			}
		}

		lengthBytes := (endByte - startByte) + 1
		ctx.Header("Content-Type", "video/mp4")
		ctx.Header("Accept-Ranges", "bytes")
		ctx.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		ctx.Header("Content-Encoding", "none")
		ctx.Header("Content-Length", strconv.Itoa(lengthBytes))
		ctx.Header("Last-Modified", streamer.Modified().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT"))
		ctx.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%s", startByte, endByte, rangeTotalStr))
		ctx.Status(http.StatusPartialContent)

		if ctx.Request.Method != http.MethodHead {
			_, err := streamer.Seek(int64(startByte), io.SeekStart)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			_, err = io.CopyN(ctx.Writer, streamer, int64(lengthBytes))
		}
	} else {
		ctx.File(types.SERV.FileTree.Get(m.GetFiles()[0]).GetAbsPath())
	}
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

	if errors.Is(err, dataStore.ErrNoCache) {
		f := types.SERV.FileTree.Get(m.GetFiles()[0])
		if f != nil {
			types.SERV.TaskDispatcher.ScanDirectory(f.GetParent(), types.SERV.Caster)
			ctx.Status(http.StatusNoContent)
			return
		} else {
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
		ctx.Status(http.StatusInternalServerError)
		return
	}
}

func hideMedia(ctx *gin.Context) {
	body, err := readCtxBody[mediaIdsBody](ctx)
	if err != nil {
		return
	}

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
		err := m.Hide()
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
