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
	"github.com/go-chi/chi/v5"
)

// GetMedia godoc
//
//	@ID			GetMedia
//
//	@Summary	Get paginated media
//	@Tags		Media
//	@Produce	json
//	@Param		raw			query		bool				false	"Include raw files"		Enums(true, false)	default(false)
//	@Param		hidden		query		bool				false	"Include hidden media"	Enums(true, false)	default(false)
//	@Param		sort		query		string				false	"Sort by field"			Enums(createDate)	default(createDate)
//	@Param		page		query		int					false	"Page of medias to get"
//	@Param		limit		query		int					false	"Number of medias to get"
//	@Param		folderIds	query		string				false	"Search only in given folders"			SchemaExample([fId1, fId2])
//	@Param		mediaIds	query		string				false	"Get only media with the provided ids"	SchemaExample([mId1, id2])
//	@Success	200			{object}	rest.MediaBatchInfo	"Media Batch"
//	@Success	400
//	@Success	500
//	@Router		/media [get]
func getMediaBatch(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(w, r)
	if SafeErrorAndExit(err, w) {
		return
	}

	folderIdsStr := r.URL.Query().Get("folderIds")
	if folderIdsStr != "" {
		var folderIds []fileTree.FileId
		err := json.Unmarshal([]byte(folderIdsStr), &folderIds)
		if SafeErrorAndExit(err, w) {
			return
		}

		getMediaInFolders(pack, u, folderIds, w, r)
		return
	}

	mediaIdsStr := r.URL.Query().Get("mediaIds")
	if mediaIdsStr != "" {
		var mediaIds []fileTree.FileId
		err := json.Unmarshal([]byte(mediaIdsStr), &mediaIds)
		if SafeErrorAndExit(err, w) {
			return
		}

		var medias []*models.Media
		for _, mId := range mediaIds {
			medias = append(medias, pack.MediaService.Get(mId))
		}

		getMediaInFolders(pack, u, mediaIds, w, r)
		writeJson(w, http.StatusOK, rest.MediaBatchInfo{Media: medias})
		return
	}

	sort := r.URL.Query().Get("sort")
	if sort == "" {
		sort = "createDate"
	}

	raw := r.URL.Query().Get("raw") == "true"
	hidden := r.URL.Query().Get("hidden") == "true"

	var page int64
	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		page, err = strconv.ParseInt(pageStr, 10, 32)
		if SafeErrorAndExit(err, w) {
			return
		}
	} else {
		page = 0
	}

	var limit int64
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		limit, err = strconv.ParseInt(limitStr, 10, 32)
		if SafeErrorAndExit(err, w) {
			return
		}
	} else {
		limit = 100
	}

	var albumFilter []models.AlbumId
	albumsStr := r.URL.Query().Get("albums")
	if albumsStr != "" {
		err = json.Unmarshal([]byte(albumsStr), &albumFilter)
		if SafeErrorAndExit(err, w) {
			return
		}
	}

	var mediaFilter []models.ContentId
	for _, albumId := range albumFilter {
		a := pack.AlbumService.Get(albumId)
		mediaFilter = append(mediaFilter, a.GetMedias()...)
	}

	ms, err := pack.MediaService.GetFilteredMedia(u, sort, 1, mediaFilter, raw, hidden)
	if SafeErrorAndExit(err, w) {
		return
	}

	var slicedMs []*models.Media
	if (page+1)*limit > int64(len(ms)) {
		slicedMs = ms[(page)*limit:]
	} else {
		slicedMs = ms[(page)*limit : (page+1)*limit]
	}

	res := rest.MediaBatchInfo{Media: slicedMs, MediaCount: len(ms)}

	writeJson(w, http.StatusOK, res)
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
func getMediaTypes(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	writeJson(w, http.StatusOK, pack.MediaService.GetMediaTypes())
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
func getMediaInfo(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	mediaId := chi.URLParam(r, "mediaId")
	m := pack.MediaService.Get(mediaId)
	if m == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	writeJson(w, http.StatusOK, m)
}

// GetMediaImage godoc
//
//	@Id			GetMediaImage
//
//	@Summary	Get a media image bytes
//	@Tags		Media
//	@Produce	image/webp, image/png, image/jpeg
//	@Param		mediaId		path		string	true	"Media Id"
//	@Param		extension	path		string	true	"Extension"
//	@Param		quality		query		string	true	"Image Quality"	Enums(thumbnail, fullres)
//	@Param		page		query		int		false	"Page number"
//	@Success	200			{string}	binary	"image bytes"
//	@Success	500
//	@Router		/media/{mediaId}.{extension} [get]
func getMediaImage(w http.ResponseWriter, r *http.Request) {
	quality := models.MediaQuality(r.URL.Query().Get("quality"))
	format := chi.URLParam(r, "extension")
	getProcessedMedia(quality, format, w, r)
}

func streamVideo(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	// u, err := getUserFromCtx(w, r)
	// if SafeErrorAndExit(err, w) {
	// 	return
	// }

	sh, err := getShareFromCtx[*models.FileShare](w, r)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	mediaId := chi.URLParam(r, "mediaId")
	m := pack.MediaService.Get(mediaId)
	if m == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if !pack.MediaService.GetMediaType(m).IsVideo() {
		writeJson(w, http.StatusBadRequest, gin.H{"error": "media is not of type video"})
		return
	}

	streamer, err := pack.MediaService.StreamVideo(m, pack.UserService.GetRootUser(), sh)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	chunkName := chi.URLParam(r, "chunkName")
	if chunkName != "" {
		http.ServeFile(w, r, streamer.GetEncodeDir()+chunkName)
		return
	}

	playlistFilePath := filepath.Join(streamer.GetEncodeDir(), "list.m3u8")
	for {
		_, err := os.Stat(playlistFilePath)
		if streamer.Err() != nil {
			log.ShowErr(streamer.Err())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err != nil {
			time.Sleep(time.Millisecond * 100)
		} else {
			break
		}
	}
	http.ServeFile(w, r, playlistFilePath)
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
func hideMedia(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	body, err := readCtxBody[rest.MediaIdsParams](w, r)
	if err != nil {
		return
	}

	hidden := r.URL.Query().Get("hidden") == "true"

	medias := make([]*models.Media, len(body.MediaIds))
	for i, mId := range body.MediaIds {
		m := pack.MediaService.Get(mId)
		if m == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		medias[i] = m
	}

	for _, m := range medias {
		err = pack.MediaService.HideMedia(m, hidden)
		if err != nil {
			log.ShowErr(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func adjustMediaDate(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	body, err := readCtxBody[rest.MediaTimeBody](w, r)
	if err != nil {
		return
	}

	anchor := pack.MediaService.Get(body.AnchorId)
	if anchor == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	extras := internal.Map(
		body.MediaIds, func(mId models.ContentId) *models.Media { return pack.MediaService.Get(body.AnchorId) },
	)

	err = pack.MediaService.AdjustMediaDates(anchor, body.NewTime, extras)
	if err != nil {
		log.ShowErr(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func likeMedia(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(w, r)
	if SafeErrorAndExit(err, w) {
		return
	}
	if u == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	mediaId := chi.URLParam(r, "mediaId")
	liked := r.URL.Query().Get("liked") == "true"

	err = pack.MediaService.SetMediaLiked(mediaId, liked, u.GetUsername())
	if err != nil {
		log.ShowErr(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Helper function
func getMediaInFolders(pack *models.ServicePack, u *models.User, folderIds []string, w http.ResponseWriter, r *http.Request) {
	var folders []*fileTree.WeblensFileImpl
	for _, folderId := range folderIds {
		f, err := pack.FileService.GetFileSafe(folderId, u, nil)
		if err != nil {
			log.ShowErr(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		folders = append(folders, f)
	}

	res := rest.MediaBatchInfo{Media: pack.MediaService.RecursiveGetMedia(folders...)}

	writeJson(w, http.StatusOK, res)
}

// Helper function
func getProcessedMedia(q models.MediaQuality, format string, w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(w, r)
	if SafeErrorAndExit(err, w) {
		return
	}
	mediaId := chi.URLParam(r, "mediaId")

	m := pack.MediaService.Get(mediaId)
	if m == nil {
		writeJson(w, http.StatusNotFound, gin.H{"error": "Media with given ID not found"})
		return
	}

	var pageNum int
	if q == models.HighRes && m.PageCount > 1 {
		pageString := r.URL.Query().Get("page")
		pageNum, err = strconv.Atoi(pageString)
		if err != nil {
			log.Debug.Println("Bad page number trying to get fullres multi-page image")
			writeJson(w, http.StatusBadRequest, gin.H{"error": "bad page number"})
			return
		}
	}

	if q == models.Video && !pack.MediaService.GetMediaType(m).IsVideo() {
		writeJson(w, http.StatusBadRequest, gin.H{"error": "media type is not video"})
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
				writeJson(w, http.StatusInternalServerError, gin.H{"error": "Failed to launch process media task"})
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		} else {
			log.ErrTrace(err)
			writeJson(w, http.StatusInternalServerError, gin.H{"error": "Failed to get media content"})
			return
		}
	}

	if err != nil {
		log.ErrTrace(err)
		writeJson(w, http.StatusInternalServerError, gin.H{"error": "Failed to get media content"})
		return
	}

	if format == "png" {
		image := bimg.NewImage(bs)
		bs, err = image.Convert(bimg.PNG)
		if err != nil {
			log.ErrTrace(err)
			writeJson(w, http.StatusInternalServerError, gin.H{"error": "Failed to convert image to PNG"})
			return
		}
	}

	w.Header().Set("Cache-Control", "max-age=3600")

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bs)

	if err != nil {
		log.ErrTrace(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
