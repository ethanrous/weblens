package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/rest"
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
//	@Param		search		query		string				false	"Search string"
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
	u, err := getUserFromCtx(r, true)
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

		getMediaInFolders(pack, u, folderIds, w)
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

		getMediaInFolders(pack, u, mediaIds, w)
		batch := rest.NewMediaBatchInfo(medias)
		writeJson(w, http.StatusOK, batch)
		return
	}

	sort := r.URL.Query().Get("sort")
	if sort == "" {
		sort = "createDate"
	}

	raw := r.URL.Query().Get("raw") == "true"
	hidden := r.URL.Query().Get("hidden") == "true"
	search := r.URL.Query().Get("search")

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

	ms, err := pack.MediaService.GetFilteredMedia(u, sort, 1, mediaFilter, raw, hidden, search)
	if SafeErrorAndExit(err, w) {
		return
	}

	var slicedMs []*models.Media
	if (page+1)*limit > int64(len(ms)) {
		slicedMs = ms[(page)*limit:]
	} else {
		slicedMs = ms[(page)*limit : (page+1)*limit]
	}

	batch := rest.NewMediaBatchInfo(slicedMs)

	writeJson(w, http.StatusOK, batch)
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
	mimeMap, extMap := pack.MediaService.GetMediaTypes().GetMaps()
	typeInfo := rest.MediaTypeInfo{
		MimeMap: mimeMap,
		ExtMap:  extMap,
	}
	w.Header().Add("Cache-Control", "max-age=3600")
	writeJson(w, http.StatusOK, typeInfo)
}

// CleanupMedia godoc
//
//	@ID			CleanupMedia
//
//	@Security	SessionAuth[admin]
//	@Security	ApiKeyAuth[admin]
//
//	@Summary	Make sure all media is correctly synced with the file system
//	@Tags		Media
//	@Produce	json
//	@Success	200
//	@Failure	500
//	@Router		/media/cleanup  [post]
func cleanupMedia(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	err := pack.MediaService.Cleanup()
	if SafeErrorAndExit(err, w) {
		return
	}
	w.WriteHeader(http.StatusOK)
}

// DropMedia godoc
//
//	@ID			DropMedia
//
//	@Security	SessionAuth[admin]
//	@Security	ApiKeyAuth[admin]
//
//	@Summary	DANGEROUS. Drop all computed media and clear thumbnail in-memory and filesystem cache. Must be server owner.
//	@Tags		Media
//	@Produce	json
//	@Success	200
//	@Failure	403
//	@Failure	500
//	@Router		/media/drop  [post]
func dropMedia(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r, false)
	if SafeErrorAndExit(err, w) {
		return
	}

	log.Debug.Println(u.Username, u.IsOwner())
	if !u.IsOwner() {
		SafeErrorAndExit(werror.ErrNotOwner, w)
		return
	}

	err = pack.MediaService.Drop()
	if SafeErrorAndExit(err, w) {
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetMediaInfo godoc
//
//	@ID			GetMediaInfo
//
//	@Summary	Get media info
//	@Tags		Media
//	@Produce	json
//	@Param		mediaId	path		string			true	"Media Id"
//	@Success	200		{object}	rest.MediaInfo	"Media Info"
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
//	@Produce	image/*
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

	pack.Log.Debug.Func(func(l log.Logger) { l.Println("Streaming video", chi.URLParam(r, "chunkName")) })

	sh, err := getShareFromCtx[*models.FileShare](w, r)
	if err != nil {
		return
	}

	mediaId := chi.URLParam(r, "mediaId")
	m := pack.MediaService.Get(mediaId)
	if m == nil {
		writeError(w, http.StatusNotFound, werror.ErrNoMedia)
		return
	} else if !pack.MediaService.GetMediaType(m).Video {
		writeError(w, http.StatusBadRequest, werror.Errorf("media is not of type video"))
		return
	}

	streamer, err := pack.MediaService.StreamVideo(m, pack.UserService.GetRootUser(), sh)
	if SafeErrorAndExit(err, w) {
		return
	}

	chunkName := chi.URLParam(r, "chunkName")
	if chunkName != "" {
		chunkFile, err := streamer.GetChunk(chunkName)
		if SafeErrorAndExit(err, w) {
			return
		}

		pack.Log.Trace.Println("Serving chunk", chunkName)
		wrote, err := io.Copy(w, chunkFile)
		if SafeErrorAndExit(err, w) {
			return
		}

		err = chunkFile.Close()
		pack.Log.Trace.Println("Chunk [%s] wrote [%d] bytes", chunkName, wrote)
		if SafeErrorAndExit(err, w) {
			return
		}
		return
	}

	listFile, err := streamer.GetListFile()
	if SafeErrorAndExit(err, w) {
		return
	}

	// if !bytes.HasSuffix(listFile, []byte("#EXT-X-ENDLIST")) {
	// 	listFile = append(listFile, []byte("#EXT-X-ENDLIST")...)
	// }

	_, err = w.Write(listFile)
	SafeErrorAndExit(err, w)
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
			pack.Log.ShowErr(err)
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
		pack.Log.ShowErr(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// SetMediaLiked godoc
//
//	@Id			SetMediaLiked
//
//	@Security	SessionAuth
//
//	@Summary	Like a media
//	@Tags		Media
//	@Produce	json
//	@Param		mediaId	path	string	true	"Id of media"
//	@Param		shareId	query	string	false	"ShareId"
//	@Param		liked	query	bool	true	"Liked status to set"
//	@Success	200
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/media/{mediaId}/liked [patch]
func setMediaLiked(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r, true)
	if SafeErrorAndExit(err, w) {
		return
	}

	mediaId := chi.URLParam(r, "mediaId")
	liked := r.URL.Query().Get("liked") == "true"

	err = pack.MediaService.SetMediaLiked(mediaId, liked, u.GetUsername())
	if SafeErrorAndExit(err, w) {
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetMediaFile godoc
//
//	@Id			GetMediaFile
//
//	@Security	SessionAuth
//	@Security	ApiKeyAuth
//
//	@Summary	Get file of media by id
//	@Tags		Media
//	@Produce	json
//	@Param		mediaId	path		string			true	"Id of media"
//	@Success	200		{object}	rest.FileInfo	"File info of file media was created from"
//	@Success	404
//	@Success	500
//	@Router		/media/{mediaId}/file [get]
func getMediaFile(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r, true)
	if SafeErrorAndExit(err, w) {
		return
	}

	mediaId := chi.URLParam(r, "mediaId")

	m := pack.MediaService.Get(mediaId)
	if m == nil {
		SafeErrorAndExit(werror.ErrNoMedia, w)
		return
	}

	var f *fileTree.WeblensFileImpl
	for _, fId := range m.GetFiles() {
		fu, err := pack.FileService.GetFileSafe(fId, u, nil)
		if err == nil && fu.GetPortablePath().RootName() == "USERS" {
			break
		}
	}

	if f == nil {
		f, err = pack.FileService.GetFileByContentId(m.ID())
		if SafeErrorAndExit(err, w) {
			return
		}
	}

	fInfo, err := rest.WeblensFileToFileInfo(f, pack, false)
	if SafeErrorAndExit(err, w) {
		return
	}
	writeJson(w, http.StatusOK, fInfo)
}

// Helper function
func getMediaInFolders(pack *models.ServicePack, u *models.User, folderIds []string, w http.ResponseWriter) {
	var folders []*fileTree.WeblensFileImpl
	for _, folderId := range folderIds {
		f, err := pack.FileService.GetFileSafe(folderId, u, nil)
		if err != nil {
			pack.Log.ShowErr(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		folders = append(folders, f)
	}

	ms := pack.MediaService.RecursiveGetMedia(folders...)
	batch := rest.NewMediaBatchInfo(ms)

	writeJson(w, http.StatusOK, batch)
}

// Helper function
func getProcessedMedia(q models.MediaQuality, format string, w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r, true)
	if SafeErrorAndExit(err, w) {
		return
	}
	mediaId := chi.URLParam(r, "mediaId")

	m := pack.MediaService.Get(mediaId)
	if m == nil {
		writeJson(w, http.StatusNotFound, rest.WeblensErrorInfo{Error: "Media with given ID not found"})
		return
	}

	var pageNum int
	if q == models.HighRes && m.PageCount > 1 {
		pageString := r.URL.Query().Get("page")
		pageNum, err = strconv.Atoi(pageString)
		if err != nil {
			pack.Log.Debug.Println("Bad page number trying to get fullres multi-page image")
			writeJson(w, http.StatusBadRequest, rest.WeblensErrorInfo{Error: "bad page number"})
			return
		}
	}

	if q == models.Video && !pack.MediaService.GetMediaType(m).Video {
		writeJson(w, http.StatusBadRequest, rest.WeblensErrorInfo{Error: "media type is not video"})
		return
	}

	bs, err := pack.MediaService.FetchCacheImg(m, q, pageNum)

	if errors.Is(err, werror.ErrNoCache) {
		files := m.GetFiles()
		f, err := pack.FileService.GetFileSafe(files[len(files)-1], u, nil)
		if SafeErrorAndExit(err, w) {
			return
		}

		meta := models.ScanMeta{
			File:         f.GetParent(),
			FileService:  pack.FileService,
			MediaService: pack.MediaService,
			TaskService:  pack.TaskService,
			TaskSubber:   pack.ClientService,
		}
		_, err = pack.TaskService.DispatchJob(models.ScanDirectoryTask, meta, nil)
		if SafeErrorAndExit(err, w) {
			return
		}
		pack.Log.Debug.Printf("Image %s has no cache", m.ID())
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if SafeErrorAndExit(err, w) {
		return
	}

	// Instruct the client to cache images that are returned
	w.Header().Set("Cache-Control", "max-age=3600")
	w.Header().Set("Content-Type", "image/"+format)

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bs)

	if err != nil {
		pack.Log.ErrTrace(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
