package media

import (
	"io"
	"mime"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/ethanrous/weblens/docs"
	file_model "github.com/ethanrous/weblens/models/file"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/models/share"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/net"
	"github.com/ethanrous/weblens/modules/slices"
	"github.com/ethanrous/weblens/modules/structs"
	file_api "github.com/ethanrous/weblens/routers/api/v1/file"
	"github.com/ethanrous/weblens/services/context"
	media_service "github.com/ethanrous/weblens/services/media"
	"github.com/ethanrous/weblens/services/reshape"
)

// GetMedia godoc
//
//	@ID			GetMedia
//
//	@Summary	Get paginated media
//	@Tags		Media
//	@Produce	json
//	@Param		request	body		structs.MediaBatchParams	true	"Media Batch Params"
//	@Param		shareId		query		string					false	"File ShareId"
//	@Success	200			{object}	structs.MediaBatchInfo	"Media Batch"
//	@Success	400
//	@Success	500
//	@Router		/media [post]
func GetMediaBatch(ctx context.RequestContext) {
	reqParams, err := net.ReadRequestBody[structs.MediaBatchParams](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusBadRequest, err)

		return
	}

	if len(reqParams.FolderIds) != 0 {
		for _, folderId := range reqParams.FolderIds {
			_, err := file_api.CheckFileAccessById(ctx, folderId, share.SharePermissionView)
			if err != nil {
				ctx.Error(http.StatusForbidden, err)

				return
			}
		}

		if reqParams.SortDirection == 0 {
			reqParams.SortDirection = -1
		}

		page := reqParams.Page
		limit := reqParams.Limit

		if reqParams.Search != "" {
			page = 0
			limit = 9999999
		}

		media, err := getMediaInFolders(ctx, reqParams.FolderIds, limit, page, reqParams.SortDirection, reqParams.Raw)
		if err != nil {
			ctx.Log().Error().Stack().Err(err).Msg("Failed to get media in folders")
			ctx.Error(http.StatusInternalServerError, err)
		}

		if reqParams.Search != "" {
			scoredMedia, err := media_service.SortMediaByTextSimilarity(ctx.AppContext, reqParams.Search, media, 0.22)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err)

				return
			}

			media = slices.Map(scoredMedia, func(m media_service.MediaWithScore) *media_model.Media { return m.Media })
		}

		batch := reshape.NewMediaBatchInfo(media)
		ctx.JSON(http.StatusOK, batch)

		return
	}

	if len(reqParams.MediaIds) != 0 {
		var medias []*media_model.Media

		for _, mId := range reqParams.MediaIds {
			m, err := media_model.GetMediaByContentId(ctx, mId)
			if err != nil {
				ctx.Log().Error().Stack().Err(err).Msg("Failed to get media by id")
				ctx.Error(http.StatusInternalServerError, err)

				return
			}

			medias = append(medias, m)
		}

		batch := reshape.NewMediaBatchInfo(medias)
		ctx.JSON(http.StatusOK, batch)

		return
	}

	if reqParams.Sort == "" {
		reqParams.Sort = "createDate"
	}

	var mediaFilter []media_model.ContentId

	ms, err := media_model.GetMedia(ctx, ctx.Requester.Username, reqParams.Sort, 1, mediaFilter, reqParams.Raw, reqParams.Hidden, reqParams.Search)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	var slicedMs []*media_model.Media
	if (reqParams.Page+1)*reqParams.Limit > len(ms) {
		slicedMs = ms[(reqParams.Page)*reqParams.Limit:]
	} else {
		slicedMs = ms[(reqParams.Page)*reqParams.Limit : (reqParams.Page+1)*reqParams.Limit]
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
//	@Success	200	{object}	structs.MediaTypesInfo	"Media types"
//	@Router		/media/types  [get]
func GetMediaTypes(ctx context.RequestContext) {
	mime, ext := media_model.GetMaps()

	mimeInfo := make(map[string]structs.MediaTypeInfo)
	for k, v := range mime {
		mimeInfo[k] = reshape.MediaTypeToMediaTypeInfo(v)
	}

	extInfo := make(map[string]structs.MediaTypeInfo)
	for k, v := range ext {
		extInfo[k] = reshape.MediaTypeToMediaTypeInfo(v)
	}

	typesInfo := structs.MediaTypesInfo{
		MimeMap: mimeInfo,
		ExtMap:  extInfo,
	}
	ctx.JSON(http.StatusOK, typesInfo)
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
func CleanupMedia(ctx context.RequestContext) {
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
func DropMedia(ctx context.RequestContext) {
	err := media_model.DropMediaCollection(ctx)
	if err != nil {
		ctx.Log().Error().Stack().Err(err).Msg("Failed to drop media collection")
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	for _, p := range ctx.GetCache("photoCache").ScanKeys() {
		ctx.GetCache("photoCache").Delete(p)
	}

	err = os.RemoveAll(file_model.ThumbsDirPath.ToAbsolute())
	if err != nil {
		ctx.Error(http.StatusInternalServerError, errors.Errorf("Failed to remove thumbnails directory: %w", err))

		return
	}

	err = os.Mkdir(file_model.ThumbsDirPath.ToAbsolute(), 0755)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, errors.Errorf("Failed to re-create thumbnails directory: %w", err))

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
func GetMediaInfo(ctx context.RequestContext) {
	mediaId := ctx.Path("mediaId")

	m, err := media_model.GetMediaByContentId(ctx, mediaId)
	if err != nil {
		ctx.Log().Error().Stack().Err(err).Msg("Failed to get media by id")
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
func GetMediaImage(ctx context.RequestContext) {
	quality, ok := media_model.CheckMediaQuality(ctx.Query("quality"))
	if !ok {
		ctx.Error(http.StatusBadRequest, errors.New("Invalid quality parameter"))

		return
	}

	format := ctx.Path("extension")
	getProcessedMedia(ctx, quality, format)
}

func streamVideo(ctx context.RequestContext) {
	chunkName := ctx.Path("chunkName")

	ctx.Log().Debug().Msgf("Streaming video %s", chunkName)

	mediaId := ctx.Path("mediaId")

	m, err := media_model.GetMediaByContentId(ctx, mediaId)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)

		return
	} else if !media_model.ParseMime(m.MimeType).IsVideo {
		ctx.Error(http.StatusBadRequest, errors.New("media is not of type video"))

		return
	}

	streamer, err := media_service.StreamVideo(ctx, m)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	if chunkName != "" {
		ctx.SetContentType(mime.TypeByExtension(".mp4"))

		// Set headers to ensure caching, but require revalidation
		ctx.W.Header().Add("Cache-Control", "no-cache")

		modified, err := streamer.GetChunkModified(chunkName)
		if err == nil {
			modified = modified.Truncate(time.Second)
			ctx.SetLastModified(modified)

			modifiedSince, hasModifiedHeader := ctx.IfModifiedSince()

			// If the client has sent an If-Modified-Since header, check if the chunk has been modified since then
			if hasModifiedHeader && (modified.Equal(modifiedSince) || modified.Truncate(time.Second).Before(modifiedSince)) {
				ctx.Status(http.StatusNotModified)

				return
			}
		} else if !errors.Is(err, media_model.ErrChunkNotFound) {
			ctx.Error(http.StatusInternalServerError, err)

			return
		}

		chunkFile, err := streamer.GetChunk(chunkName)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)

			return
		}

		defer chunkFile.Close()

		_, err = io.Copy(ctx, chunkFile)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)
		}

		return
	}

	listFile, modified, err := streamer.GetListFile()
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.SetContentType(mime.TypeByExtension(".m3u8"))
	ctx.W.Header().Add("Cache-Control", "no-cache")

	modified = modified.Truncate(time.Second)
	ctx.SetLastModified(modified)
	modifiedSince, hasModifiedHeader := ctx.IfModifiedSince()

	if hasModifiedHeader && (modified.Equal(modifiedSince) || modified.Truncate(time.Second).Before(modifiedSince)) {
		ctx.Status(http.StatusNotModified)

		return
	}

	_, err = ctx.Write(listFile)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}
}

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
func HideMedia(ctx context.RequestContext) {
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

func AdjustMediaDate(ctx context.RequestContext) {
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
func SetMediaLiked(ctx context.RequestContext) {
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
func GetMediaFile(ctx context.RequestContext) {
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
func StreamVideo(ctx context.RequestContext) {
	streamVideo(ctx)
}

// GetRandomMedia() godoc
//
//	@Id			GetRandomMedia
//
//	@Summary	Get random media
//	@Tags		Media
//	@Produce	json
//	@Param		count	query		number					true	"Number of random medias to get"
//	@Success	200		{object}	structs.MediaBatchInfo	"Media Batch"
//	@Success	404
//	@Success	500
//	@Router		/media/random [get]
func GetRandomMedia(ctx context.RequestContext) {
	countStr := ctx.Query("count")

	count, err := strconv.Atoi(countStr)
	if err != nil {
		ctx.Error(http.StatusBadRequest, err)

		return
	}

	username := ctx.AttemptGetUsername()

	if username == "" {
		ctx.Error(http.StatusUnauthorized, errors.New("unauthorized: no username provided"))

		return
	}

	medias, err := media_model.GetRandomMedias(ctx, media_model.RandomMediaOptions{Owner: username, Count: count, NoRaws: true})
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	batch := reshape.NewMediaBatchInfo(medias)
	ctx.JSON(http.StatusOK, batch)
}

// Helper function
func getMediaInFolders(ctx context.RequestContext, folderIds []string, limit, page, sortDirection int, includeRaw bool) ([]*media_model.Media, error) {
	allContentIds := []string{}

	for _, folderId := range folderIds {
		folder, err := ctx.FileService.GetFileById(ctx, folderId)

		if err != nil {
			return nil, err
		}

		err = folder.RecursiveMap(func(wfi *file_model.WeblensFileImpl) error {
			if wfi.IsDir() {
				_, err := ctx.FileService.GetChildren(ctx, wfi)

				return err
			}

			return nil
		})
		if err != nil {
			return nil, err
		}

		err = folder.RecursiveMap(func(wfi *file_model.WeblensFileImpl) error {
			if wfi.IsDir() {
				return nil
			}

			allContentIds = append(allContentIds, wfi.GetContentId())

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	medias, err := media_model.GetMediasByContentIds(ctx, limit, page, sortDirection, includeRaw, allContentIds...)
	if err != nil {
		return nil, err
	}

	return medias, nil
}

// Helper function
func getProcessedMedia(ctx context.RequestContext, q media_model.MediaQuality, format string) {
	mediaId := ctx.Path("mediaId")

	m, err := media_model.GetMediaByContentId(ctx, mediaId)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)

		return
	}

	var pageNum int

	if q == media_model.HighRes && m.PageCount > 1 {
		pageString := ctx.Query("page")

		pageNum, err = strconv.Atoi(pageString)
		if err != nil {
			ctx.Error(http.StatusBadRequest, err)

			return
		}
	}

	mt := media_model.ParseMime(m.MimeType)

	if format == "pdf" && q == media_model.HighRes && mt.IsMultiPage() {
		f, err := ctx.FileService.GetFileByContentId(ctx, m.ContentID)

		if err != nil {
			ctx.Error(http.StatusNotFound, err)

			return
		}

		pdfBytes, err := f.ReadAll()
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)

			return
		}

		_, err = ctx.Write(pdfBytes)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)
		}

		return
	}

	if q == media_model.Video && mt.IsVideo {
		ctx.Error(http.StatusBadRequest, errors.New("media type is not video"))

		return
	}

	bs, err := media_service.FetchCacheImg(ctx.AppContext, m, q, pageNum)

	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)
	}

	// Instruct the client to cache images that are returned
	ctx.SetHeader("Cache-Control", "max-age=3600")
	ctx.SetHeader("Content-Type", "image/"+format)
	ctx.Bytes(http.StatusOK, bs)
}
