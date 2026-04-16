// Package media provides handlers for media-related API endpoints.
package media

import (
	"io"
	"mime"
	"net/http"
	"slices"
	"strconv"
	"time"

	_ "github.com/ethanrous/weblens/docs" // Required for swagger docs generation
	file_model "github.com/ethanrous/weblens/models/file"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/models/share"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/modules/wlslices"
	"github.com/ethanrous/weblens/modules/wlstructs"
	file_api "github.com/ethanrous/weblens/routers/api/v1/file"
	"github.com/ethanrous/weblens/services/ctxservice"
	file_service "github.com/ethanrous/weblens/services/file"
	media_service "github.com/ethanrous/weblens/services/media"
	"github.com/ethanrous/weblens/services/reshape"
)

const (
	maxSearchResults  = 9999999
	maxRecursionDepth = 100
)

// GetMediaBatch godoc
//
//	@ID			GetMedia
//
//	@Summary	Get paginated media
//	@Tags		Media
//	@Produce	json
//	@Param		shareID			query		string						false	"File ShareID"
//	@Param		raw				query		bool						false	"Include raw media"		Enums(true, false)
//	@Param		hidden			query		bool						false	"Include hidden media"	Enums(true, false)
//	@Param		sort			query		string						false	"Sort field"			Enums(createDate)
//	@Param		sortDirection	query		int							false	"Sort direction"		Enums(1, -1)
//	@Param		search			query		string						false	"Search query"
//	@Param		page			query		int							false	"Page number"
//	@Param		limit			query		int							false	"Page size"
//	@Param		folderIDs		query		[]string					false	"Folder IDs to filter by"
//	@Param		mediaIDs		query		[]string					false	"Media IDs to fetch"
//	@Success	200				{object}	wlstructs.MediaBatchInfo	"Media Batch"
//	@Success	400
//	@Success	500
//	@Router		/media [get]
func GetMediaBatch(ctx ctxservice.RequestContext) {
	raw := ctx.QueryBool("raw")
	hidden := ctx.QueryBool("hidden")

	sort := ctx.Query("sort")
	if sort == "" {
		sort = "createDate"
	} else if slices.Index([]string{"createDate"}, sort) == -1 {
		ctx.Error(http.StatusBadRequest, wlerrors.Errorf("Invalid sort field '%s'", sort))

		return
	}

	search := ctx.Query("search")
	folderIDs := ctx.QueryArray("folderIDs")
	mediaIDs := ctx.QueryArray("mediaIDs")

	sortDirection, err := ctx.QueryInt("sortDirection")
	if err != nil {
		ctx.Error(http.StatusBadRequest, err)

		return
	}

	page, err := ctx.QueryInt("page")
	if err != nil {
		ctx.Error(http.StatusBadRequest, err)

		return
	} else if page < 0 {
		ctx.Error(http.StatusBadRequest, wlerrors.New("Page number cannot be negative"))

		return
	}

	limit, err := ctx.QueryInt("limit")
	if err != nil {
		ctx.Error(http.StatusBadRequest, err)

		return
	} else if limit <= 0 {
		ctx.Error(http.StatusBadRequest, wlerrors.New("Limit must be greater than 0"))

		return
	}

	if len(folderIDs) != 0 {
		getMediaByFolders(ctx, folderIDs, search, int(sortDirection), int(page), int(limit), raw)

		return
	}

	if len(mediaIDs) != 0 {
		getMediaByIDs(ctx, mediaIDs)

		return
	}

	// The paginated fallback uses ctx.Requester.Username and is not share-aware.
	// Require authentication when no folder or media IDs are provided.
	if !ctx.IsLoggedIn {
		ctx.Error(http.StatusUnauthorized, wlerrors.New("authentication required"))

		return
	}

	getMediaPaginated(ctx, sort, raw, hidden, int(page), int(limit))
}

func getMediaByFolders(ctx ctxservice.RequestContext, folderIDs []string, search string, sortDirection, page, limit int, raw bool) {
	for _, folderID := range folderIDs {
		_, err := file_api.CheckFileAccessByID(ctx, folderID, share.SharePermissionViewMedia)
		if err != nil {
			return
		}
	}

	if sortDirection == 0 {
		sortDirection = -1
	}

	if search != "" {
		page = 0
		limit = maxSearchResults
	}

	media, totalMediaCount, err := getMediaInFolders(ctx, folderIDs, limit, page, sortDirection, raw)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	if totalMediaCount == 0 {
		batch := reshape.NewMediaBatchInfo(nil)
		batch.TotalMediaCount = 0
		ctx.JSON(http.StatusOK, batch)

		return
	}

	ctx.Log().Trace().Msgf("Got %d media in folder(s) %+v", len(media), folderIDs)

	scores := []float64{}

	if search != "" {
		scoredMedia, err := media_service.SortMediaByTextSimilarity(ctx.AppContext, search, media, 0.60)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)

			return
		}

		media = wlslices.Map(scoredMedia, func(m media_service.ScoreWrapper) *media_model.Media { return m.Media })
		totalMediaCount = len(media)

		if totalMediaCount == 0 {
			batch := reshape.NewMediaBatchInfo(nil)
			batch.TotalMediaCount = 0
			ctx.JSON(http.StatusOK, batch)

			return
		}

		scores = wlslices.Map(scoredMedia, func(m media_service.ScoreWrapper) float64 { return m.Score })

		ctx.Log().Trace().Msgf("Search query '%s' got %d matching scored media", search, len(scores))
	}

	batch := reshape.NewMediaBatchInfo(media, reshape.MediaBatchOptions{Scores: scores})
	batch.TotalMediaCount = totalMediaCount
	ctx.JSON(http.StatusOK, batch)
}

func getMediaByIDs(ctx ctxservice.RequestContext, mediaIDs []string) {
	var medias []*media_model.Media

	for _, mID := range mediaIDs {
		m, err := media_model.GetMediaByContentID(ctx, mID)
		if err != nil {
			ctx.Log().Error().Stack().Err(err).Msg("Failed to get media by id")
			ctx.Error(http.StatusInternalServerError, err)

			return
		}

		// Verify the requester has access to view media of at least one backing file
		hasAccess := false

		for _, fileID := range m.GetFiles() {
			if _, err := file_api.CheckFileAccessByID(ctx, fileID, share.SharePermissionViewMedia); err == nil {
				hasAccess = true

				break
			}
		}

		if !hasAccess {
			ctx.Error(http.StatusForbidden, wlerrors.New("not authorized to access this media"))

			return
		}

		medias = append(medias, m)
	}

	batch := reshape.NewMediaBatchInfo(medias)
	batch.TotalMediaCount = len(medias)
	ctx.JSON(http.StatusOK, batch)
}

func getMediaPaginated(ctx ctxservice.RequestContext, sort string, raw, hidden bool, page, limit int) {
	var mediaFilter []media_model.ContentID

	ms, err := media_model.GetMedia(ctx, ctx.Requester.Username, sort, 1, mediaFilter, raw, hidden)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	start := page * limit
	if start > len(ms) {
		start = len(ms)
	}

	end := start + limit
	if end > len(ms) {
		end = len(ms)
	}

	batch := reshape.NewMediaBatchInfo(ms[start:end])

	ctx.JSON(http.StatusOK, batch)
}

// GetMediaTypes godoc
//
//	@ID			GetMediaTypes
//
//	@Summary	Get media type dictionary
//	@Tags		Media
//	@Produce	json
//	@Success	200	{object}	wlstructs.MediaTypesInfo	"Media types"
//	@Router		/media/types  [get]
func GetMediaTypes(ctx ctxservice.RequestContext) {
	mime, ext := media_model.GetMaps()

	mimeInfo := make(map[string]wlstructs.MediaTypeInfo)
	for k, v := range mime {
		mimeInfo[k] = reshape.MediaTypeToMediaTypeInfo(v)
	}

	extInfo := make(map[string]wlstructs.MediaTypeInfo)
	for k, v := range ext {
		extInfo[k] = reshape.MediaTypeToMediaTypeInfo(v)
	}

	typesInfo := wlstructs.MediaTypesInfo{
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
func CleanupMedia(ctx ctxservice.RequestContext) {
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
//	@Param		username	query	string	false	"Username of owner whose media to drop. If empty, drops all media."
//	@Produce	json
//	@Success	200
//	@Failure	403
//	@Failure	500
//	@Router		/media/drop  [post]
func DropMedia(ctx ctxservice.RequestContext) {
	username := ctx.Query("username")

	removedIDs, err := media_model.DropMediaByOwner(ctx, username)
	if err != nil {
		ctx.Log().Error().Stack().Err(err).Msg("Failed to drop media collection")
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	mediaCache := ctx.GetCache("photoCache")
	for _, p := range mediaCache.ScanKeys() {
		mediaCache.Delete(p)
	}

	ctx.Log().Debug().Msgf("Removing media cache files: %+v", removedIDs)

	err = file_service.RemoveCacheFilesWithFilter(ctx, removedIDs)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, wlerrors.Errorf("Failed to remove media cache files: %w", err))
	}

	ctx.Status(http.StatusOK)
}

// DropHDIRs godoc
//
//	@ID			DropHDIRs
//
//	@Security	SessionAuth[admin]
//	@Security	ApiKeyAuth[admin]
//
//	@Summary	Drop all computed media HDIR data. Must be server owner.
//	@Tags		Media
//	@Produce	json
//	@Success	200
//	@Failure	403
//	@Failure	500
//	@Router		/media/drop/hdirs  [post]
func DropHDIRs(ctx ctxservice.RequestContext) {
	err := media_model.DropHDIRs(ctx)
	if err != nil {
		ctx.Log().Error().Stack().Err(err).Msg("Failed to drop media hdir data")
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
//	@Param		mediaID	path		string				true	"Media ID"
//	@Success	200		{object}	wlstructs.MediaInfo	"Media Info"
//	@Router		/media/{mediaID}/info [get]
func GetMediaInfo(ctx ctxservice.RequestContext) {
	mediaID := ctx.Path("mediaID")

	m, err := media_model.GetMediaByContentID(ctx, mediaID)
	if err != nil {
		ctx.Log().Error().Stack().Err(err).Msg("Failed to get media by id")
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	// Cache for 1 hour
	ctx.SetHeader("Cache-Control", "max-age=3600")
	ctx.JSON(http.StatusOK, reshape.MediaToMediaInfo(m))
}

// GetMediaImage godoc
//
//	@ID			GetMediaImage
//
//	@Summary	Get a media image bytes
//	@Tags		Media
//	@Produce	image/*
//	@Param		mediaID		path		string	true	"Media ID"
//	@Param		extension	path		string	true	"Extension"
//	@Param		quality		query		string	true	"Image Quality"	Enums(thumbnail, fullres)
//	@Param		page		query		int		false	"Page number"
//	@Param		shareID		query		string	false	"Share ID"
//	@Success	200			{string}	binary	"image bytes"
//	@Success	500
//	@Router		/media/{mediaID}.{extension} [get]
func GetMediaImage(ctx ctxservice.RequestContext) {
	quality, ok := media_model.CheckMediaQuality(ctx.Query("quality"))
	if !ok {
		ctx.Error(http.StatusBadRequest, wlerrors.New("Invalid quality parameter"))

		return
	}

	format := ctx.Path("extension")
	getProcessedMedia(ctx, quality, format)
}

func streamVideo(ctx ctxservice.RequestContext) {
	chunkName := ctx.Path("chunkName")

	ctx.Log().Debug().Msgf("Streaming video %s", chunkName)

	mediaID := ctx.Path("mediaID")

	m, err := media_model.GetMediaByContentID(ctx, mediaID)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)

		return
	} else if !media_model.ParseMime(m.MimeType).IsVideo {
		ctx.Error(http.StatusBadRequest, wlerrors.New("media is not of type video"))

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
		} else if !wlerrors.Is(err, media_model.ErrChunkNotFound) {
			ctx.Error(http.StatusInternalServerError, err)

			return
		}

		chunkFile, err := streamer.GetChunk(chunkName)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)

			return
		}

		defer chunkFile.Close() //nolint:errcheck

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

// HideMedia godoc
//
//	@ID			SetMediaVisibility
//
//	@Summary	Set media visibility
//	@Tags		Media
//	@Produce	json
//	@Param		hidden		query	bool						true	"Set the media visibility"	Enums(true, false)
//	@Param		mediaIDs	body	wlstructs.MediaIDsParams	true	"MediaIDs to change visibility of"
//	@Success	200
//	@Success	404
//	@Success	500
//	@Router		/media/visibility [patch]
func HideMedia(ctx ctxservice.RequestContext) {
	ctx.Status(http.StatusNotImplemented)
	// pack := getServices(r)
	// log := hlog.FromRequest(r)
	// body, err := readCtxBody[structs.MediaIDsParams](w, r)
	// if err != nil {
	// 	return
	// }
	//
	// hidden := r.URL.Query().Get("hidden") == "true"
	//
	// medias := make([]*models.Media, len(body.MediaIDs))
	// for i, mID := range body.MediaIDs {
	// 	m := pack.MediaService.Get(mID)
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
	// w.WriteHeader(http.StatusOK)
	_ = ""
}

// AdjustMediaDate adjusts the date metadata for a media item.
func AdjustMediaDate(ctx ctxservice.RequestContext) {
	ctx.Status(http.StatusNotImplemented)

	// pack := getServices(r)
	// log := hlog.FromRequest(r)
	// body, err := readCtxBody[structs.MediaTimeBody](w, r)
	// if err != nil {
	// 	return
	// }
	//
	// anchor := pack.MediaService.Get(body.AnchorID)
	// if anchor == nil {
	// 	w.WriteHeader(http.StatusNotFound)
	// 	return
	// }
	// extras := internal.Map(
	// 	body.MediaIDs, func(mID models.ContentID) *models.Media { return pack.MediaService.Get(body.AnchorID) },
	// )
	//
	// err = pack.MediaService.AdjustMediaDates(anchor, body.NewTime, extras)
	// if err != nil {
	// 	log.Error().Stack().Err(err).Msg("")
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }
	// w.WriteHeader(http.StatusOK)
	_ = ""
}

// SetMediaLiked godoc
//
//	@ID			SetMediaLiked
//
//	@Security	SessionAuth
//
//	@Summary	Like a media
//	@Tags		Media
//	@Produce	json
//	@Param		mediaID	path	string	true	"ID of media"
//	@Param		shareID	query	string	false	"ShareID"
//	@Param		liked	query	bool	true	"Liked status to set"
//	@Success	200
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/media/{mediaID}/liked [patch]
func SetMediaLiked(ctx ctxservice.RequestContext) {
	ctx.Status(http.StatusNotImplemented)
}

// GetMediaFile godoc
//
//	@ID			GetMediaFile
//
//	@Security	SessionAuth
//	@Security	ApiKeyAuth
//
//	@Summary	Get file of media by id
//	@Tags		Media
//	@Produce	json
//	@Param		mediaID	path		string				true	"ID of media"
//	@Success	200		{object}	wlstructs.FileInfo	"File info of file media was created from"
//	@Success	404
//	@Success	500
//	@Router		/media/{mediaID}/file [get]
func GetMediaFile(ctx ctxservice.RequestContext) {
	ctx.Status(http.StatusNotImplemented)
}

// StreamVideo godoc
//
//	@ID			StreamVideo
//
//	@Security	SessionAuth
//	@Security	ApiKeyAuth
//
//	@Summary	Stream a video
//	@Tags		Media
//	@Produce	octet-stream
//	@Param		mediaID	path	string	true	"ID of media"
//	@Success	200
//	@Success	404
//	@Success	500
//	@Router		/media/{mediaID}/video [get]
func StreamVideo(ctx ctxservice.RequestContext) {
	streamVideo(ctx)
}

// GetRandomMedia godoc
//
//	@ID			GetRandomMedia
//
//	@Summary	Get random media
//	@Tags		Media
//	@Produce	json
//	@Param		count	query		number						true	"Number of random medias to get"
//	@Success	200		{object}	wlstructs.MediaBatchInfo	"Media Batch"
//	@Success	404
//	@Success	500
//	@Router		/media/random [get]
func GetRandomMedia(ctx ctxservice.RequestContext) {
	countStr := ctx.Query("count")

	count, err := strconv.Atoi(countStr)
	if err != nil {
		ctx.Error(http.StatusBadRequest, err)

		return
	}

	username := ctx.Requester.GetUsername()

	medias, err := media_model.GetRandomMedias(ctx, media_model.RandomMediaOptions{Owner: username, Count: count, NoRaws: true})
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	batch := reshape.NewMediaBatchInfo(medias)
	ctx.JSON(http.StatusOK, batch)
}

// Helper function
func getMediaInFolders(ctx ctxservice.RequestContext, folderIDs []string, limit, page, sortDirection int, includeRaw bool) ([]*media_model.Media, int, error) {
	allContentIDs := []string{}

	for _, folderID := range folderIDs {
		folder, err := ctx.FileService.GetFileByID(ctx, folderID)
		if err != nil {
			return nil, -1, err
		}

		contentIDs, err := collectContentIDs(ctx, folder)
		if err != nil {
			return nil, -1, err
		}

		allContentIDs = append(allContentIDs, contentIDs...)
	}

	medias, err := media_model.GetPagedMedias(ctx, limit, page, sortDirection, includeRaw, allContentIDs...)
	if err != nil {
		return nil, -1, err
	}

	return medias, len(allContentIDs), nil
}

// collectContentIDs recursively walks a folder tree, loading children and
// collecting content IDs from non-directory files. Directories named
// ".user_trash" are skipped so that trashed files do not appear in the timeline.
// Recursion is bounded by maxRecursionDepth to prevent stack overflow on
// deeply nested or cyclic directory structures.
func collectContentIDs(ctx ctxservice.RequestContext, folder *file_model.WeblensFileImpl) ([]string, error) {
	return collectContentIDsWithDepth(ctx, folder, 0)
}

func collectContentIDsWithDepth(ctx ctxservice.RequestContext, folder *file_model.WeblensFileImpl, depth int) ([]string, error) {
	if !folder.IsDir() {
		return []string{folder.GetContentID()}, nil
	}

	if folder.Name() == file_model.UserTrashDirName {
		return nil, nil
	}

	if depth >= maxRecursionDepth {
		ctx.Log().Warn().Msgf("Max recursion depth reached at folder %s", folder.Name())

		return nil, nil
	}

	children, err := ctx.FileService.GetChildren(ctx, folder)
	if err != nil {
		return nil, err
	}

	var contentIDs []string

	for _, child := range children {
		childIDs, err := collectContentIDsWithDepth(ctx, child, depth+1)
		if err != nil {
			return nil, err
		}

		contentIDs = append(contentIDs, childIDs...)
	}

	return contentIDs, nil
}

// getProcessedMedia retrieves and writes the processed media bytes for a given media ID, quality, and format.
// It handles both image and video media types, as well as multi-page documents like PDFs.
func getProcessedMedia(ctx ctxservice.RequestContext, q media_model.Quality, format string) {
	mediaID := ctx.Path("mediaID")

	m, err := media_model.GetMediaByContentID(ctx, mediaID)
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
	if q == media_model.HighRes && mt.IsMultiPage() && slices.Contains([]string{"pdf", "gif"}, format) {
		f, err := ctx.FileService.GetFileByContentID(ctx, m.ContentID)
		if err != nil {
			ctx.Error(http.StatusNotFound, err)

			return
		}

		_, err = file_api.CheckFileAccessByID(ctx, f.ID(), share.SharePermissionViewMedia)
		if err != nil {
			ctx.Error(http.StatusForbidden, wlerrors.New("not authorized to access this media"))

			return
		}

		multiPageBytes, err := f.ReadAll()
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)

			return
		}

		ctx.SetHeader("Cache-Control", "max-age=3600")
		ctx.SetHeader("Content-Type", "image/"+format)
		ctx.Bytes(http.StatusOK, multiPageBytes)

		return
	}

	if q == media_model.Video && mt.IsVideo {
		ctx.Error(http.StatusBadRequest, wlerrors.New("media type is not video"))

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
