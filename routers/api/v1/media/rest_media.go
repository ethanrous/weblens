package media

import (
	"encoding/json"
	"net/http"
	"strconv"

	_ "github.com/ethanrous/weblens/docs"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/reshape"
	"github.com/pkg/errors"
)

// GetMedia godoc
//
//	@ID			GetMedia
//
//	@Summary	Get paginated media
//	@Tags		Media
//	@Produce	json
//	@Param		raw			query		bool					false	"Include raw files"		Enums(true, false)	default(false)
//	@Param		hidden		query		bool					false	"Include hidden media"	Enums(true, false)	default(false)
//	@Param		sort		query		string					false	"Sort by field"			Enums(createDate)	default(createDate)
//	@Param		search		query		string					false	"Search string"
//	@Param		page		query		int						false	"Page of medias to get"
//	@Param		limit		query		int						false	"Number of medias to get"
//	@Param		folderIds	query		string					false	"Search only in given folders"			SchemaExample([fId1, fId2])
//	@Param		mediaIds	query		string					false	"Get only media with the provided ids"	SchemaExample([mId1, id2])
//	@Success	200			{object}	structs.MediaBatchInfo	"Media Batch"
//	@Success	400
//	@Success	500
//	@Router		/media [get]
func GetMediaBatch(ctx *context.RequestContext) {
	folderIdsStr := ctx.Query("folderIds")
	if folderIdsStr != "" {
		var folderIds []string
		err := json.Unmarshal([]byte(folderIdsStr), &folderIds)
		if err != nil {
			ctx.Logger.Error().Stack().Err(err).Msg("Failed to unmarshal folderIds")
			ctx.Error(http.StatusBadRequest, errors.New("Invalid folderIds format"))
			return
		}

		media, err := getMediaInFolders(ctx, folderIds)
		if err != nil {
			ctx.Logger.Error().Stack().Err(err).Msg("Failed to get media in folders")
			ctx.Error(http.StatusInternalServerError, err)
		}

		batch := reshape.NewMediaBatchInfo(media)
		ctx.JSON(http.StatusOK, batch)

		return
	}

	mediaIdsStr := ctx.Query("mediaIds")
	if mediaIdsStr != "" {
		var mediaIds []string
		err := json.Unmarshal([]byte(mediaIdsStr), &mediaIds)
		if err != nil {
			ctx.Logger.Error().Stack().Err(err).Msg("Failed to unmarshal mediaIds")
			ctx.Error(http.StatusBadRequest, errors.New("Invalid mediaIds format"))
			return
		}

		var medias []*media_model.Media
		for _, mId := range mediaIds {
			m, err := media_model.GetMediaById(ctx, mId)
			if err != nil {
				ctx.Logger.Error().Stack().Err(err).Msg("Failed to get media by id")
				ctx.Error(http.StatusInternalServerError, err)
				return
			}
			medias = append(medias, m)
		}

		batch := reshape.NewMediaBatchInfo(medias)
		ctx.JSON(http.StatusOK, batch)
		return
	}

	sort := ctx.Query("sort")
	if sort == "" {
		sort = "createDate"
	}

	raw := ctx.Query("raw") == "true"
	hidden := ctx.Query("hidden") == "true"
	search := ctx.Query("search")

	var page int64
	var err error
	pageStr := ctx.Query("page")
	if pageStr != "" {
		page, err = strconv.ParseInt(pageStr, 10, 32)
		if err != nil {
			ctx.Logger.Error().Stack().Err(err).Msg("Failed to parse page number")
			ctx.Error(http.StatusBadRequest, errors.New("Invalid page number"))
			return
		}
	} else {
		page = 0
	}

	var limit int64
	limitStr := ctx.Query("limit")
	if limitStr != "" {
		limit, err = strconv.ParseInt(limitStr, 10, 32)
		if err != nil {
			ctx.Logger.Error().Stack().Err(err).Msg("Failed to parse limit number")
			ctx.Error(http.StatusBadRequest, errors.New("Invalid limit number"))
			return
		}
	} else {
		limit = 100
	}

	var mediaFilter []media_model.ContentId
	ms, err := media_model.GetMedia(ctx, ctx.Requester.Username, sort, 1, mediaFilter, raw, hidden, search)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)
		return
	}

	var slicedMs []*media_model.Media
	if (page+1)*limit > int64(len(ms)) {
		slicedMs = ms[(page)*limit:]
	} else {
		slicedMs = ms[(page)*limit : (page+1)*limit]
	}

	batch := reshape.NewMediaBatchInfo(slicedMs)

	ctx.JSON(http.StatusOK, batch)
}

// GetMediaTypes godoc
//
//	@ID			GetMediaTypes
//
//	@Summary	Get media type dictionary
//	@Tags		Media
//	@Produce	json
//	@Success	200	{object}	structs.MediaTypeInfo	"Media types"
//	@Router		/media/types  [get]
func GetMediaTypes(ctx *context.RequestContext) {
	ctx.W.Write([]byte(media_model.MediaTypeJson))
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
func CleanupMedia(ctx *context.RequestContext) {
	// pack := getServices(r)
	// log := hlog.FromRequest(r)
	// err := pack.MediaService.Cleanup()
	// if SafeErrorAndExit(err, w, log) {
	// 	return
	// }
	ctx.Status(http.StatusNotImplemented)
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
func DropMedia(ctx *context.RequestContext) {
	err := media_model.DropMediaCollection(ctx)
	if err != nil {
		ctx.Logger.Error().Stack().Err(err).Msg("Failed to drop media collection")
		ctx.Error(http.StatusInternalServerError, err)
		return
	}
	ctx.Status(http.StatusOK)
}

// GetMediaInfo godoc
//
//	@ID			GetMediaInfo
//
//	@Summary	Get media info
//	@Tags		Media
//	@Produce	json
//	@Param		mediaId	path		string				true	"Media Id"
//	@Success	200		{object}	structs.MediaInfo	"Media Info"
//	@Router		/media/{mediaId}/info [get]
func GetMediaInfo(ctx *context.RequestContext) {
	mediaId := ctx.Path("mediaId")
	m, err := media_model.GetMediaById(ctx, mediaId)
	if err != nil {
		ctx.Logger.Error().Stack().Err(err).Msg("Failed to get media by id")
		ctx.Error(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, reshape.MediaToMediaInfo(m))
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
func GetMediaImage(ctx *context.RequestContext) {
	quality, ok := media_model.CheckMediaQuality(ctx.Query("quality"))
	if !ok {
		ctx.Error(http.StatusBadRequest, errors.New("Invalid quality parameter"))
		return
	}
	format := ctx.Path("extension")
	getProcessedMedia(ctx, quality, format)
}

// func streamVideo(w http.ResponseWriter, r *http.Request) {
// 	pack := getServices(r)
// 	log := hlog.FromRequest(r)
//
// 	log.Debug().Func(func(e *zerolog.Event) { e.Msgf("Streaming video %s", chi.URLParam(r, "chunkName")) })
//
// 	sh, err := getShareFromCtx[*models.FileShare](w, r)
// 	if err != nil {
// 		return
// 	}
//
// 	mediaId := chi.URLParam(r, "mediaId")
// 	m := pack.MediaService.Get(mediaId)
// 	if m == nil {
// 		writeError(w, http.StatusNotFound, werror.ErrNoMedia)
// 		return
// 	} else if !pack.MediaService.GetMediaType(m).Video {
// 		writeError(w, http.StatusBadRequest, werror.Errorf("media is not of type video"))
// 		return
// 	}
//
// 	streamer, err := pack.MediaService.StreamVideo(m, pack.UserService.GetRootUser(), sh)
// 	if SafeErrorAndExit(err, w, log) {
// 		return
// 	}
//
// 	chunkName := chi.URLParam(r, "chunkName")
// 	if chunkName != "" {
// 		chunkFile, err := streamer.GetChunk(chunkName)
// 		if SafeErrorAndExit(err, w, log) {
// 			return
// 		}
//
// 		log.Trace().Msgf("Serving chunk %s", chunkName)
// 		wrote, err := io.Copy(w, chunkFile)
// 		if SafeErrorAndExit(err, w, log) {
// 			return
// 		}
//
// 		err = chunkFile.Close()
// 		log.Trace().Msgf("Chunk [%s] wrote [%d] bytes", chunkName, wrote)
// 		if SafeErrorAndExit(err, w, log) {
// 			return
// 		}
// 		return
// 	}
//
// 	listFile, err := streamer.GetListFile()
// 	if SafeErrorAndExit(err, w, log) {
// 		return
// 	}
//
// 	// if !bytes.HasSuffix(listFile, []byte("#EXT-X-ENDLIST")) {
// 	// 	listFile = append(listFile, []byte("#EXT-X-ENDLIST")...)
// 	// }
//
// 	_, err = w.Write(listFile)
// 	SafeErrorAndExit(err, w, log)
// }

// SetMediaVisibility godoc
//
//	@Id			SetMediaVisibility
//
//	@Summary	Set media visibility
//	@Tags		Media
//	@Produce	json
//	@Param		hidden		query	bool					true	"Set the media visibility"	Enums(true, false)
//	@Param		mediaIds	body	structs.MediaIdsParams	true	"MediaIds to change visibility of"
//	@Success	200
//	@Success	404
//	@Success	500
//	@Router		/media/visibility [patch]
func HideMedia(ctx *context.RequestContext) {
	ctx.Status(http.StatusNotImplemented)
	// pack := getServices(r)
	// log := hlog.FromRequest(r)
	// body, err := readCtxBody[structs.MediaIdsParams](w, r)
	// if err != nil {
	// 	return
	// }
	//
	// hidden := r.URL.Query().Get("hidden") == "true"
	//
	// medias := make([]*models.Media, len(body.MediaIds))
	// for i, mId := range body.MediaIds {
	// 	m := pack.MediaService.Get(mId)
	// 	if m == nil {
	// 		w.WriteHeader(http.StatusNotFound)
	// 		return
	// 	}
	// 	medias[i] = m
	// }
	//
	// for _, m := range medias {
	// 	err = pack.MediaService.HideMedia(m, hidden)
	// 	if err != nil {
	// 		log.Error().Stack().Err(err).Msg("")
	// 		w.WriteHeader(http.StatusInternalServerError)
	// 		return
	// 	}
	// }
	//
	// w.WriteHeader(http.StatusOK)
}

func AdjustMediaDate(ctx *context.RequestContext) {
	ctx.Status(http.StatusNotImplemented)

	// pack := getServices(r)
	// log := hlog.FromRequest(r)
	// body, err := readCtxBody[structs.MediaTimeBody](w, r)
	// if err != nil {
	// 	return
	// }
	//
	// anchor := pack.MediaService.Get(body.AnchorId)
	// if anchor == nil {
	// 	w.WriteHeader(http.StatusNotFound)
	// 	return
	// }
	// extras := internal.Map(
	// 	body.MediaIds, func(mId models.ContentId) *models.Media { return pack.MediaService.Get(body.AnchorId) },
	// )
	//
	// err = pack.MediaService.AdjustMediaDates(anchor, body.NewTime, extras)
	// if err != nil {
	// 	log.Error().Stack().Err(err).Msg("")
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }
	//
	// w.WriteHeader(http.StatusOK)
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
func SetMediaLiked(ctx *context.RequestContext) {
	ctx.Status(http.StatusNotImplemented)

	// pack := getServices(r)
	// log := hlog.FromRequest(r)
	// u, err := getUserFromCtx(r, true)
	// if SafeErrorAndExit(err, w, log) {
	// 	return
	// }
	//
	// mediaId := chi.URLParam(r, "mediaId")
	// liked := r.URL.Query().Get("liked") == "true"
	//
	// err = pack.MediaService.SetMediaLiked(mediaId, liked, u.GetUsername())
	// if SafeErrorAndExit(err, w, log) {
	// 	return
	// }
	//
	// w.WriteHeader(http.StatusOK)
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
//	@Param		mediaId	path		string				true	"Id of media"
//	@Success	200		{object}	structs.FileInfo	"File info of file media was created from"
//	@Success	404
//	@Success	500
//	@Router		/media/{mediaId}/file [get]
func GetMediaFile(ctx *context.RequestContext) {
	ctx.Status(http.StatusNotImplemented)

	// pack := getServices(r)
	// log := hlog.FromRequest(r)
	// u, err := getUserFromCtx(r, true)
	// if SafeErrorAndExit(err, w, log) {
	// 	return
	// }
	//
	// mediaId := chi.URLParam(r, "mediaId")
	//
	// m := pack.MediaService.Get(mediaId)
	// if m == nil {
	// 	SafeErrorAndExit(werror.ErrNoMedia, w)
	// 	return
	// }
	//
	// var f *fileTree.WeblensFileImpl
	// for _, fId := range m.GetFiles() {
	// 	fu, err := pack.FileService.GetFileSafe(fId, u, nil)
	// 	if err == nil && fu.GetPortablePath().RootName() == "USERS" {
	// 		break
	// 	}
	// }
	//
	// if f == nil {
	// 	f, err = pack.FileService.GetFileByContentId(m.ID())
	// 	if SafeErrorAndExit(err, w, log) {
	// 		return
	// 	}
	// }
	//
	// fInfo, err := structs.WeblensFileToFileInfo(f, pack, false)
	// if SafeErrorAndExit(err, w, log) {
	// 	return
	// }
	// writeJson(w, http.StatusOK, fInfo)
}

// StreamVideo godoc
//
//	@Id			StreamVideo
//
//	@Security	SessionAuth
//	@Security	ApiKeyAuth
//
//	@Summary	Stream a video
//	@Tags		Media
//	@Produce	octet-stream
//	@Param		mediaId	path	string	true	"Id of media"
//	@Success	200
//	@Success	404
//	@Success	500
//	@Router		/media/{mediaId}/video [get]
func StreamVideo(ctx *context.RequestContext) {
	ctx.Status(http.StatusNotImplemented)

	// pack := getServices(r)
	// log := hlog.FromRequest(r)
	// u, err := getUserFromCtx(r, true)
	// if SafeErrorAndExit(err, w, log) {
	// 	return
	// }
	//
	// streamVideo(w, r)
}

// Helper function
func getMediaInFolders(ctx *context.RequestContext, folderIds []string) ([]*media_model.Media, error) {
	// var folders []*fileTree.WeblensFileImpl
	// for _, folderId := range folderIds {
	// 	f, err := pack.FileService.GetFileSafe(folderId, u, nil)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	folders = append(folders, f)
	// }
	//
	// ms := pack.MediaService.RecursiveGetMedia(folders...)
	// batch := structs.NewMediaBatchInfo(ms)
	//
	// writeJson(w, http.StatusOK, batch)

	ctx.Status(http.StatusNotImplemented)
	return nil, nil
}

// Helper function
func getProcessedMedia(ctx *context.RequestContext, q media_model.MediaQuality, format string) {
	ctx.Status(http.StatusNotImplemented)

	// mediaId := chi.URLParam(r, "mediaId")
	//
	// m := pack.MediaService.Get(mediaId)
	// if m == nil {
	// 	writeJson(w, http.StatusNotFound, structs.WeblensErrorInfo{Error: "Media with given ID not found"})
	// 	return
	// }
	//
	// var pageNum int
	// if q == models.HighRes && m.PageCount > 1 {
	// 	pageString := r.URL.Query().Get("page")
	// 	pageNum, err = strconv.Atoi(pageString)
	// 	if err != nil {
	// 		log.Debug().Func(func(e *zerolog.Event) { e.Msgf("Bad page number trying to get fullres multi-page image") })
	// 		writeJson(w, http.StatusBadRequest, structs.WeblensErrorInfo{Error: "bad page number"})
	// 		return
	// 	}
	// }
	//
	// if q == models.Video && !pack.MediaService.GetMediaType(m).Video {
	// 	writeJson(w, http.StatusBadRequest, structs.WeblensErrorInfo{Error: "media type is not video"})
	// 	return
	// }
	//
	// bs, err := pack.MediaService.FetchCacheImg(m, q, pageNum)
	//
	// if errors.Is(err, werror.ErrNoCache) {
	// 	files := m.GetFiles()
	// 	f, err := pack.FileService.GetFileSafe(files[len(files)-1], u, nil)
	// 	if SafeErrorAndExit(err, w, log) {
	// 		return
	// 	}
	//
	// 	meta := models.ScanMeta{
	// 		File:         f.GetParent(),
	// 		FileService:  pack.FileService,
	// 		MediaService: pack.MediaService,
	// 		TaskService:  pack.TaskService,
	// 		TaskSubber:   pack.ClientService,
	// 	}
	// 	_, err = pack.TaskService.DispatchJob(models.ScanDirectoryTask, meta, nil)
	// 	if SafeErrorAndExit(err, w, log) {
	// 		return
	// 	}
	// 	log.Debug().Func(func(e *zerolog.Event) { e.Msgf("Image %s has no cache", m.ID()) })
	// 	w.WriteHeader(http.StatusNoContent)
	// 	return
	// }
	//
	// if SafeErrorAndExit(err, w, log) {
	// 	return
	// }
	//
	// // Instruct the client to cache images that are returned
	// w.Header().Set("Cache-Control", "max-age=3600")
	// w.Header().Set("Content-Type", "image/"+format)
	//
	// w.WriteHeader(http.StatusOK)
	// _, err = w.Write(bs)
	//
	// if err != nil {
	// 	log.Error().Stack().Err(err).Msg("")
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
}
