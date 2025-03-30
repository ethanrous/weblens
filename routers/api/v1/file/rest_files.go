package file

import (
	"bytes"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	cover_model "github.com/ethanrous/weblens/models/cover"
	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/job"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/models/rest"
	share_model "github.com/ethanrous/weblens/models/share"
	"github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/net"
	"github.com/ethanrous/weblens/modules/structs"
	task_mod "github.com/ethanrous/weblens/modules/task"
	"github.com/ethanrous/weblens/services/auth"
	"github.com/ethanrous/weblens/services/context"
	"github.com/ethanrous/weblens/services/file"
	"github.com/ethanrous/weblens/services/journal"
	media_service "github.com/ethanrous/weblens/services/media"
	"github.com/ethanrous/weblens/services/reshape"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func checkFileAccess(ctx *context.RequestContext) (file *file_model.WeblensFileImpl, err error) {
	fileId := ctx.Path("fileId")
	if fileId == "" {
		fileId = ctx.Path("folderId")
	}

	// Check if the request is a takeout request (zip file)
	isTakeout := ctx.Query("isTakeout")
	if isTakeout == "true" {
		file, err = ctx.FileService.GetZip(fileId)
		if err != nil {
			ctx.Error(http.StatusNotFound, err)
			return
		}
	}

	return checkFileAccessById(ctx, fileId)
}

func checkFileAccessById(ctx *context.RequestContext, fileId string) (file *file_model.WeblensFileImpl, err error) {
	file, err = ctx.FileService.GetFileById(fileId)
	if err != nil {
		// Handle error if file not found
		if errors.Is(err, file_model.ErrFileNotFound) {
			ctx.Error(http.StatusNotFound, err)
			return
		}
		// For any other error
		ctx.Error(http.StatusInternalServerError, err)
		return
	}

	// Check if the user has access to the file
	if !auth.CanUserAccessFile(ctx.Requester, file, ctx.Share) {
		// If the user does not have access, return Unauthorized
		err = errors.New("access denied to file")
		ctx.Error(http.StatusUnauthorized, err)
		return
	}

	return
}

// GetFile godoc
//
//	@ID	GetFile
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Get information about a file
//	@Tags		Files
//	@Produce	json
//	@Param		fileId	path		string			true	"File Id"
//	@Param		shareId	query		string			false	"Share Id"
//	@Success	200		{object}	rest.FileInfo	"File Info"
//	@Failure	401
//	@Failure	404
//	@Router		/files/{fileId} [get]
func GetFile(ctx *context.RequestContext) {
	file, err := checkFileAccess(ctx)
	if err != nil {
		return
	}

	fileInfo, err := reshape.WeblensFileToFileInfo(ctx, file, false)
	ctx.JSON(http.StatusOK, fileInfo)
}

// GetFileText godoc
//
//	@ID	GetFileText
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Get the text of a text file
//	@Tags		Files
//	@Produce	plain
//	@Param		fileId	path		string	true	"File Id"
//	@Param		shareId	query		string	false	"Share Id"
//	@Success	200		{string}	string	"File text"
//	@Failure	400
//	@Router		/files/{fileId}/text [get]
func GetFileText(ctx *context.RequestContext) {
	file, err := checkFileAccess(ctx)
	if err != nil {
		return
	}

	filename := file.GetPortablePath().Filename()

	dotIndex := strings.LastIndex(filename, ".")
	if filename[dotIndex:] != ".txt" {
		ctx.Error(http.StatusBadRequest, errors.New("file is not a text file"))
		return
	}

	http.ServeFile(ctx.W, ctx.Req, file.GetPortablePath().ToAbsolute())
}

// GetFileStats godoc
//
//	@ID	GetFileStats
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Get the statistics of a file
//	@Tags		Files
//	@Produce	json
//	@Param		fileId	path	string	true	"File Id"
//	@Failure	400
//	@Failure	501
//	@Router		/files/{fileId}/stats [get]
func GetFileStats(ctx *context.RequestContext) {
	_, err := checkFileAccess(ctx)
	if err != nil {
		return
	}

	ctx.W.WriteHeader(http.StatusNotImplemented)
}

// DownloadFile godoc
//
//	@ID	DownloadFile
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Download a file
//	@Tags		Files
//	@Produce	octet-stream
//	@Param		fileId		path		string					true	"File Id"
//	@Param		shareId		query		string					false	"Share Id"
//	@Param		format		query		string					false	"File format conversion"
//	@Param		isTakeout	query		bool					false	"Is this a takeout file"	Enums(true, false)	default(false)
//	@Success	200			{string}	binary					"File content"
//	@Success	404			{object}	rest.WeblensErrorInfo	"Error Info"
//	@Router		/files/{fileId}/download [get]
func DownloadFile(ctx *context.RequestContext) {
	file, err := checkFileAccess(ctx)
	if err != nil {
		return
	}

	// TODO: Make sure to check if the requester is another tower
	// i := getInstanceFromCtx(r)
	//
	// if i != nil {
	// 	file, err = ctx.FileService.GetFileSafe(fileId, pack.UserService.GetRootUser(), nil)
	// 	if SafeErrorAndExit(err, w, log) {
	// 		return
	// 	}
	// }

	// TODO: Make this part of the content_type header instead of just a query param
	format := ctx.Query("format")
	if format != "" {
		m, err := media_model.GetMediaById(ctx, file.GetContentId())
		if err != nil {
			if errors.Is(err, media_model.ErrMediaNotFound) {
				ctx.Error(http.StatusNotFound, err)
				return
			}
			ctx.Error(http.StatusInternalServerError, err)
			return
		}

		convertedImg, err := media_service.GetConverted(m, format)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)
		}

		_, err = ctx.W.Write(convertedImg)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)
			return
		}
	}

	ctx.Logger.Debug().Func(func(e *zerolog.Event) { e.Msgf("Downloading file %s", file.GetPortablePath()) })

	http.ServeFile(ctx.W, ctx.Req, file.GetPortablePath().ToAbsolute())
}

// GetFolderHistory godoc
//
//	@ID			GetFolderHistory
//
//	@Security	SessionAuth
//
//	@Summary	Get actions of a folder at a given time
//	@Tags		Folder
//	@Param		fileId		path	string				true	"File Id"
//	@Param		timestamp	query	int					true	"Past timestamp to view the folder at, in ms since epoch"
//	@Success	200			{array}	rest.FileActionInfo	"File actions"
//	@Failure	400
//	@Failure	500
//	@Router		/files/{fileId}/history [get]
func GetFolderHistory(ctx *context.RequestContext) {
	file, err := checkFileAccess(ctx)
	if err != nil {
		return
	}

	pastTime := time.Now()
	milliStr := ctx.Query("timestamp")
	if milliStr != "" {
		millis, err := strconv.ParseInt(milliStr, 10, 64)
		if err != nil {
			ctx.Error(http.StatusBadRequest, errors.New("invalid timestamp format"))
			return
		}
		pastTime = time.UnixMilli(millis)
	}

	pastFile, err := journal.GetPastFile(file.ID(), pastTime)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)
		return
	}

	actions, err := journal.GetActionsByPath(ctx, pastFile.GetPortablePath())
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)
		return
	}

	actionInfos := []structs.FileActionInfo{}
	for _, a := range actions {
		actionInfos = append(actionInfos, reshape.FileActionToFileActionInfo(a))
	}

	ctx.JSON(http.StatusOK, actionInfos)
}

// SearchByFilename godoc
//
//	@ID			SearchByFilename
//
//	@Security	SessionAuth
//
//	@Summary	Search for files by filename
//	@Tags		Files
//
//	@Param		search			query	string			true	"Filename to search for"
//	@Param		baseFolderId	query	string			false	"The folder to search in, defaults to the user's home folder"
//	@Success	200				{array}	rest.FileInfo	"File Info"
//	@Failure	400
//	@Failure	401
//	@Failure	500
//	@Router		/files/search [get]
func SearchByFilename(ctx *context.RequestContext) {

	filenameSearch := ctx.Query("search")
	if filenameSearch == "" {
		ctx.Error(http.StatusBadRequest, errors.New("missing 'search' query parameter"))
		return
	}

	baseFolderId := ctx.Query("baseFolderId")
	if baseFolderId == "" {
		baseFolderId = ctx.Requester.HomeId
	}

	baseFolder, err := checkFileAccessById(ctx, baseFolderId)
	if err != nil {
		return
	}

	if !baseFolder.IsDir() {
		ctx.Error(http.StatusBadRequest, errors.New("the provided base folder ID is not a directory"))
		return
	}

	var fileIds []string
	var filenames []string
	_ = baseFolder.RecursiveMap(
		func(f *file_model.WeblensFileImpl) error {
			fileIds = append(fileIds, f.ID())
			filenames = append(filenames, f.GetPortablePath().Filename())
			return nil
		},
	)

	matches := fuzzy.RankFindFold(filenameSearch, filenames)
	slices.SortFunc(
		matches, func(a, b fuzzy.Rank) int {
			return a.Distance - b.Distance
		},
	)

	var fileInfos []structs.FileInfo
	for _, match := range matches {
		f, err := ctx.FileService.GetFileById(fileIds[match.OriginalIndex])
		if err != nil {
			ctx.Logger.Error().Stack().Err(err).Msgf("Failed to get file by ID: %s", fileIds[match.OriginalIndex])
			continue
		}

		if f.ID() == ctx.Requester.HomeId || f.ID() == ctx.Requester.TrashId {
			continue
		}

		fileInfo, err := reshape.WeblensFileToFileInfo(ctx, f, false)
		if err != nil {
			ctx.Logger.Error().Stack().Err(err).Msgf("Failed to convert file to FileInfo for file ID: %s", f.ID())
			continue
		}

		fileInfos = append(fileInfos, fileInfo)
	}

	ctx.JSON(http.StatusOK, fileInfos)
}

// CreateFolder godoc
//
//	@ID	CreateFolder
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Create a new folder
//	@Tags		Folder
//	@Accept		json
//	@Produce	json
//	@Param		request	body		rest.CreateFolderBody	true	"New folder body"
//	@Param		shareId	query		string					false	"Share Id"
//	@Success	200		{object}	rest.FileInfo			"File Info"
//	@Router		/folder [post]
func CreateFolder(ctx *context.RequestContext) {
	body, err := net.ReadRequestBody[rest.CreateFolderBody](ctx.Req)
	if err != nil {
		return
	}

	if body.NewFolderName == "" {
		ctx.Error(http.StatusBadRequest, errors.New("Missing body parameter 'newFolderName'"))
		return
	}

	parentFolder, err := checkFileAccessById(ctx, body.ParentFolderId)
	if err != nil {
		return
	}

	// var children []*file_model.WeblensFileImpl
	// if len(body.Children) != 0 {
	// 	for _, fileId := range body.Children {
	// 		child, err := ctx.FileService.GetFileSafe(fileId, u, nil)
	// 		if err != nil {
	// 			log.Error().Stack().Err(err).Msg("")
	// 			ctx.Error(http.StatusBadRequest, errors.New(err.Error()))
	// 			return
	// 		}
	// 		children = append(children, child)
	// 	}
	// }

	newDir, err := ctx.FileService.CreateFolder(parentFolder, body.NewFolderName)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)
		return
	}

	// err = ctx.FileService.MoveFiles(children, newDir, "USERS", pack.Caster)
	// if SafeErrorAndExit(err, w, log) {
	// 	return
	// }

	newDirInfo, err := reshape.WeblensFileToFileInfo(ctx, newDir, false)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, newDirInfo)
}

// GetFolder godoc
//
//	@ID	GetFolder
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Get a folder
//	@Tags		Folder
//	@Accept		json
//	@Produce	json
//	@Param		folderId	path		string					true	"Folder Id"
//	@Param		shareId		query		string					false	"Share Id"
//	@Param		timestamp	query		int						false	"Past timestamp to view the folder at, in ms since epoch"
//	@Success	200			{object}	rest.FolderInfoResponse	"Folder Info"
//	@Router		/folder/{folderId} [get]
func GetFolder(ctx *context.RequestContext) {
	folder, err := checkFileAccess(ctx)
	if err != nil {
		return
	}

	milliStr := ctx.Query("timestamp")
	date := time.UnixMilli(0)
	if milliStr != "" {
		millis, err := strconv.ParseInt(milliStr, 10, 64)
		if err != nil {
			ctx.Error(http.StatusBadRequest, errors.New("invalid timestamp format"))
			return
		}

		date = time.UnixMilli(millis)
	}

	if date.Unix() != 0 {
		formatRespondPastFolderInfo(ctx, folder.ID(), date)
		return
	}

	formatRespondFolderInfo(ctx, folder)
}

// SetFolderCover godoc
//
//	@ID			SetFolderCover
//
//	@Security	SessionAuth
//
//	@Summary	Set the cover image of a folder
//	@Tags		Folder
//	@Param		folderId	path	string	true	"Folder Id"
//	@Param		mediaId		query	string	true	"Media Id"
//	@Success	200
//	@Failure	400
//	@Failure	404
//	@Failure	500
//	@Router		/folder/{folderId}/cover [patch]
func SetFolderCover(ctx *context.RequestContext) {
	folder, err := checkFileAccess(ctx)
	if err != nil {
		return
	}

	mediaId := ctx.Query("mediaId")

	media, err := media_model.GetMediaById(ctx, mediaId)
	if err != nil {
		if errors.Is(err, media_model.ErrMediaNotFound) {
			// If the media doesn't exist, we can still set the cover to an empty state
			// This allows us to remove the cover if needed
			ctx.Error(http.StatusBadRequest, err)
			return
		}
		ctx.Error(http.StatusInternalServerError, err)
		return
	}

	_, err = cover_model.SetCoverPhoto(ctx, folder.ID(), media.ID())
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)
		return
	}

	// TODO: Add websocket notification to clients that the cover has been updated
	// pack.Caster.PushFileUpdate(folder, media)

	ctx.W.WriteHeader(http.StatusOK)
}

// ScanFolder godoc
//
//	@ID	ScanFolder
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Dispatch a folder scan
//	@Tags		Folder
//	@Param		shareId	query	string			false	"Share Id"
//	@Param		request	body	rest.ScanBody	true	"Scan parameters"
//	@Success	200
//	@Failure	404
//	@Failure	500
//	@Router		/folder/scan [post]
func ScanDir(ctx *context.RequestContext) {
	// sh, err := getShareFromCtx[*models.FileShare](ctx.Req)
	// if err != nil {
	// 	return
	// }
	//
	// body, err := io.ReadAll(r.Body)
	// if err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
	// var scanInfo rest.ScanBody
	// err = json.Unmarshal(body, &scanInfo)
	// if err != nil {
	// 	log.Error().Stack().Err(err).Msg("")
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	return
	// }
	//
	// dir, err := ctx.FileService.GetFileSafe(scanInfo.FolderId, u, sh)
	// if err != nil {
	// 	w.WriteHeader(http.StatusNotFound)
	// 	return
	// }
	//
	// meta := models.ScanMeta{
	// 	File:         dir,
	// 	FileService:  ctx.FileService,
	// 	MediaService: pack.MediaService,
	// 	TaskSubber:   pack.ClientService,
	// 	TaskService:  pack.TaskService,
	// }
	// _, err = pack.TaskService.DispatchJob(models.ScanDirectoryTask, meta, nil)
	// if SafeErrorAndExit(err, w, log) {
	// 	return
	// }
	//
	// w.WriteHeader(http.StatusOK)
}

// GetSharedFiles godoc
//
//	@ID			GetSharedFiles
//
//	@Security	SessionAuth
//
//	@Summary	Get files shared with the logged in user
//	@Tags		Files
//	@Success	200	{object}	rest.FolderInfoResponse	"All the top-level files shared with the user"
//	@Failure	404
//	@Failure	500
//	@Router		/files/shared [get]
func GetSharedFiles(ctx *context.RequestContext) {
	shares, err := share_model.GetSharedWithUser(ctx, ctx.Requester.Username)

	var children = make([]*file_model.WeblensFileImpl, 0, len(shares))
	for _, share := range shares {
		f, err := ctx.FileService.GetFileById(share.FileId)
		if err != nil {
			if errors.Is(err, file_model.ErrFileNotFound) {
				ctx.Logger.Error().Stack().Err(err).Msg("Could not find file acompanying a file share")
				continue
			}
			ctx.Error(http.StatusInternalServerError, err)
			return
		}
		children = append(children, f)
	}

	childInfos := make([]structs.FileInfo, 0, len(children))
	for _, child := range children {
		fInfo, err := reshape.WeblensFileToFileInfo(ctx, child, false)
		if err != nil {
			// If we can't convert the file to a FileInfo, log the error and continue
			ctx.Error(http.StatusInternalServerError, errors.New("failed to convert file to FileInfo"))
			return
		}
		childInfos = append(childInfos, fInfo)
	}

	var mediaInfos = make([]structs.MediaInfo, 0)
	for _, child := range children {
		var media *media_model.Media
		if child.IsDir() {
			cover, err := cover_model.GetCoverByFolderId(ctx, child.ID()) // This will return the cover media if it exists
			if errors.Is(err, media_model.ErrMediaNotFound) {
				continue
			} else if err != nil {
				// If there was an error retrieving the cover media, log the error and continue
				ctx.Logger.Error().Stack().Err(err).Msg("Failed to get cover media for folder")
				continue
			}

			media, err = media_model.GetMediaById(ctx, cover.CoverPhotoId) // This will ensure we have the media object to send back to the client
			if err != nil {
				ctx.Logger.Error().Stack().Err(err).Msg("Failed to get media for cover photo")
				continue
			}

		} else {
			media, err = media_model.GetMediaById(ctx, child.GetContentId()) // This will ensure we have the media object to send back to the client
			if err != nil {
				continue
			}
		}

		mediaInfo := reshape.MediaToMediaInfo(media) // This will convert the media object to a MediaInfo object for the response
		mediaInfos = append(mediaInfos, mediaInfo)
	}

	fakeSelfFile := structs.FileInfo{
		Id:           "shared",
		Filename:     "SHARED",
		IsDir:        true,
		PortablePath: "SHARED:",
	}

	res := structs.FolderInfoResponse{Children: childInfos, Medias: mediaInfos, Self: fakeSelfFile}

	ctx.JSON(http.StatusOK, res)
}

// CreateTakeout godoc
//
//	@ID	CreateTakeout
//
//	@Security
//	@Security		SessionAuth
//
//	@Summary		Create a zip file
//	@Description	Dispatch a task to create a zip file of the given files, or get the id of a previously created zip file if it already exists
//	@Tags			Files
//	@Param			shareId	query		string					false	"Share Id"
//	@Param			request	body		rest.FilesListParams	true	"File Ids"
//	@Success		200		{object}	rest.TakeoutInfo		"Zip Takeout Info"
//	@Success		202		{object}	rest.TakeoutInfo		"Task Dispatch Info"
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/takeout [post]
func CreateTakeout(ctx *context.RequestContext) {
	takeoutRequest, err := net.ReadRequestBody[rest.FilesListParams](ctx.Req)
	if err != nil {
		return
	}

	if len(takeoutRequest.FileIds) == 0 {
		ctx.Error(http.StatusBadRequest, errors.New("Cannot takeout 0 files"))
		return
	}

	// Make sure user has access to all requested files
	var files []*file_model.WeblensFileImpl
	for _, fileId := range takeoutRequest.FileIds {
		file, err := ctx.FileService.GetFileById(fileId)
		if err != nil {
			ctx.Error(http.StatusNotFound, err)
			return
		}

		files = append(files, file)
	}

	// If we only have 1 file, and it is not a directory, we should have requested to just
	// simply download that file on it's own, not zip it.
	if len(files) == 1 && !files[0].IsDir() {
		ctx.Error(http.StatusBadRequest, errors.New("Single non-directory file should not be zipped"))
		return
	}

	meta := job.ZipMeta{
		Files:     files,
		Requester: ctx.Requester,
		Share:     ctx.Share,
	}

	// TODO: Allow context to dispatch jobs
	t, err := ctx.TaskService.DispatchJob(ctx, job.CreateZipTask, meta, nil)

	ctx.W.WriteHeader(http.StatusNotImplemented)

	if err != nil {
		ctx.Error(http.StatusInternalServerError, errors.Wrap(err, "Failed to dispatch zip task"))
		return
	}

	completed, status := t.Status()
	if completed && status == task_mod.TaskSuccess {
		result := t.GetResult()
		res := structs.TakeoutInfo{TakeoutId: result["takeoutId"].(string), Single: false, Filename: result["filename"].(string)}
		ctx.JSON(http.StatusOK, res)
	} else {
		ctx.JSON(http.StatusAccepted, structs.TakeoutInfo{TaskId: t.Id()})
	}
}

// AutocompletePath godoc
//
//	@ID			AutocompletePath
//
//	@Security	SessionAuth
//
//	@Summary	Get path completion suggestions
//	@Tags		Files
//	@Param		searchPath	query		string	true	"Search path"
//	@Success	200			{object}	rest.FolderInfoResponse"Path info"
//	@Failure	500
//	@Router		/files/autocomplete [get]
func AutocompletePath(ctx *context.RequestContext) {

	searchPath := ctx.Query("searchPath")
	if len(searchPath) == 0 {
		ctx.Error(http.StatusBadRequest, errors.New("Missing 'searchPath' query parameter"))
		return
	}

	filepath, err := fs.ParsePortable(searchPath)
	if err != nil {
		ctx.Error(http.StatusBadRequest, err)
		return
	}

	lastSlashI := strings.LastIndex(searchPath, "/")
	childName := searchPath[lastSlashI+1:]
	searchPath = searchPath[:lastSlashI] + "/"

	folder, err := ctx.FileService.GetFileByFilepath(filepath)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)
		return
	}

	children := folder.GetChildren()
	if folder.GetParentId() == "ROOT" {
		trashIndex := slices.IndexFunc(children, func(f *file_model.WeblensFileImpl) bool {
			return f.ID() == ctx.Requester.TrashId
		})
		children = slices.Delete(children, trashIndex, trashIndex+1)
	}

	var filenames []string
	for _, child := range children {
		filenames = append(filenames, child.GetPortablePath().Filename())
	}

	matches := fuzzy.RankFindFold(childName, filenames)
	slices.SortFunc(
		matches, func(a, b fuzzy.Rank) int {
			diff := a.Distance - b.Distance
			if diff == 0 {
				return strings.Compare(a.Target, b.Target)
			} else {
				return diff
			}
		},
	)

	var childInfos []structs.FileInfo
	for _, match := range matches {
		f := children[match.OriginalIndex]

		childInfo, err := reshape.WeblensFileToFileInfo(ctx, f, false)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, errors.New("failed to convert file to FileInfo"))
			return
		}
		childInfos = append(childInfos, childInfo)
	}

	selfInfo, err := reshape.WeblensFileToFileInfo(ctx, folder, false)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, errors.New("failed to convert folder to FileInfo"))
		return
	}

	ret := structs.FolderInfoResponse{Children: childInfos, Self: selfInfo}
	ctx.JSON(http.StatusOK, ret)
}

// RestoreFiles godoc
//
//	@ID			RestoreFiles
//
//	@Security	SessionAuth
//
//	@Summary	Restore files from some time in the past
//	@Tags		Files
//	@Accept		json
//	@Produce	json
//	@Param		request	body		rest.RestoreFilesBody	true	"Restore files request body"
//	@Success	200		{object}	rest.RestoreFilesInfo	"Restore files info"
//	@Failure	400
//	@Failure	404
//	@Failure	500
//	@Router		/files/restore [post]
func RestoreFiles(ctx *context.RequestContext) {
	body, err := net.ReadRequestBody[rest.RestoreFilesBody](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, errors.New("Failed to read request body"))
	}

	if body.Timestamp == 0 {
		ctx.Error(http.StatusBadRequest, errors.New("Missing body parameter 'timestamp'"))
		return
	}
	ctx.W.WriteHeader(http.StatusNotImplemented)
	// restoreTime := time.UnixMilli(body.Timestamp)

	// parentLt := ctx.FileService.GetJournalByTree("USERS").Get(body.NewParentId)
	// if parentLt == nil {
	// 	ctx.Error(http.StatusNotFound, errors.New("Could not find new parent"))
	// 	return
	// }
	//
	// // New parent folder is the folder it was in at the time we are restoring from, if
	// // it still exists, otherwise it is the users home folder
	// var newParent *file_model.WeblensFileImpl
	// if parentLt.GetLatestAction().GetActionType() == fileTree.FileDelete {
	// 	newParent, err = ctx.FileService.GetFileSafe(u.HomeId, u, nil)
	//
	// 	// this should never error, but you never know
	// 	if SafeErrorAndExit(err, w, log) {
	// 		return
	// 	}
	// } else {
	// 	newParent, err = ctx.FileService.GetFileSafe(body.NewParentId, u, nil)
	// 	if SafeErrorAndExit(err, w, log) {
	// 		return
	// 	}
	// }
	//
	// // actions := parentLt.GetActions()
	// // for i, action := range actions {
	// // 	if action.Timestamp.After(restoreTime) && (action.ActionType != fileTree.FileSizeChange || i == len(actions)-1) {
	// // 		if i != 0 {
	// // 			restoreTime = actions[i-1].Timestamp
	// // 		}
	// // 		break
	// // 	}
	// // }
	//
	// err = ctx.FileService.RestoreFiles(body.FileIds, newParent, restoreTime, pack.Caster)
	// if SafeErrorAndExit(err, w, log) {
	// 	return
	// }
	//
	// res := rest.RestoreFilesInfo{NewParentId: newParent.ID()}
	//
	// writeJson(w, http.StatusOK, res)
}

// UpdateFile godoc
//
//	@ID	UpdateFile
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Update a File
//	@Tags		Files
//	@Accept		json
//	@Param		fileId	path	string					true	"File Id"
//	@Param		shareId	query	string					false	"Share Id"
//	@Param		request	body	rest.UpdateFileParams	true	"Update file request body"
//	@Success	200
//	@Failure	403
//	@Failure	404
//	@Failure	500
//	@Router		/files/{fileId} [patch]
func UpdateFile(ctx *context.RequestContext) {
	file, err := checkFileAccess(ctx)
	if err != nil {
		return
	}

	if ctx.FileService.IsFileInTrash(file) {
		ctx.Error(http.StatusForbidden, errors.New("cannot rename file in trash"))
		return
	}

	updateInfo, err := net.ReadRequestBody[rest.UpdateFileParams](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, errors.New("Failed to read request body"))
		return
	}

	err = ctx.FileService.RenameFile(file, updateInfo.NewName)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)
		return
	}

	ctx.W.WriteHeader(http.StatusOK)
}

// MoveFiles godoc
//
//	@ID	MoveFiles
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Move a list of files to a new parent folder
//	@Tags		Files
//	@Param		request	body	rest.MoveFilesParams	true	"Move files request body"
//	@Param		shareId	query	string					false	"Share Id"
//	@Success	200
//	@Failure	404
//	@Failure	500
//	@Router		/files [patch]
func MoveFiles(ctx *context.RequestContext) {
	filesData, err := net.ReadRequestBody[rest.MoveFilesParams](ctx.Req)
	if err != nil {
		return
	}

	newParent, err := checkFileAccessById(ctx, filesData.NewParentId)
	if err != nil {
		ctx.Error(http.StatusNotFound, errors.New("Could not find new parent folder"))
		return
	}

	if len(filesData.Files) == 0 {
		ctx.Error(http.StatusBadRequest, errors.New("No file ids provided"))
		return
	}

	if filesData.NewParentId == "" {
		ctx.Error(http.StatusBadRequest, errors.New("No parent id provided"))
		return
	}

	var files []*file_model.WeblensFileImpl
	parentId := ""
	for _, fileId := range filesData.Files {
		f, err := ctx.FileService.GetFileById(fileId)
		if err != nil {
			ctx.Error(http.StatusNotFound, errors.New("Could not find file with id "+fileId))
			return
		}

		if parentId == "" {
			parentId = f.GetParentId()
		} else if parentId != f.GetParentId() {
			ctx.Error(http.StatusBadRequest, errors.New("All files must have the same parent"))
			return
		}

		files = append(files, f)
	}

	err = ctx.FileService.MoveFiles(files, newParent, "USERS")
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)
		return
	}

	ctx.W.WriteHeader(http.StatusOK)
}

// UnTrashFiles godoc
//
//	@ID			UnTrashFiles
//
//	@Security	SessionAuth
//
//	@Summary	Move a list of files out of the trash, restoring them to where they were before
//	@Tags		Files
//	@Param		request	body	rest.FilesListParams	true	"Un-trash files request body"
//	@Success	200
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/files/untrash [patch]
func UnTrashFiles(ctx *context.RequestContext) {
	params, err := net.ReadRequestBody[rest.FilesListParams](ctx.Req)
	if err != nil {
		return
	}
	fileIds := params.FileIds

	var files []*file_model.WeblensFileImpl
	for _, fileId := range fileIds {
		file, err := ctx.FileService.GetFileById(fileId)
		if err != nil {
			ctx.Error(http.StatusNotFound, errors.New("Could not find file with id "+fileId))
			return
		}

		files = append(files, file)
	}

	err = ctx.FileService.ReturnFilesFromTrash(files)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, errors.Wrap(err, "Failed to un-trash files"))
	}

	ctx.W.WriteHeader(http.StatusOK)
}

// DeleteFiles godoc
//
//	@ID			DeleteFiles
//
//	@Security	SessionAuth
//
//	@Summary	Delete Files "permanently"
//	@Tags		Files
//	@Param		request		body	rest.FilesListParams	true	"Delete files request body"
//	@Param		ignoreTrash	query	boolean					false	"Delete files even if they are not in the trash"
//	@Success	200
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/files [delete]
func deleteFiles(ctx *context.RequestContext) {
	params, err := net.ReadRequestBody[rest.FilesListParams](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)
		return
	}

	if len(params.FileIds) == 0 {
		ctx.Error(http.StatusBadRequest, errors.New("No file ids provided"))
		return
	}

	// ignoreTrash := ctx.Query("ignore_trash") == "true"

	var files []*file_model.WeblensFileImpl
	for _, fileId := range params.FileIds {
		file, err := ctx.FileService.GetFileById(fileId)
		if err != nil {
			ctx.Error(http.StatusNotFound, err)
			return
		}

		// TODO: Check permissions on files

		files = append(files, file)
	}

	err = ctx.FileService.DeleteFiles(ctx, files, "USERS")
	if err != nil {
		ctx.Error(http.StatusInternalServerError, errors.Wrap(err, "Failed to delete files"))
		return
	}

	ctx.W.WriteHeader(http.StatusOK)
}

// StartUpload godoc
//
//	@ID	StartUpload
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Begin a new upload task
//	@Tags		Files
//	@Param		request	body		rest.NewUploadParams	true	"New upload request body"
//	@Param		shareId	query		string					false	"Share Id"
//	@Success	201		{object}	rest.NewUploadInfo		"Upload Info"
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/upload [post]
func NewUploadTask(ctx *context.RequestContext) {
	upInfo, err := net.ReadRequestBody[rest.NewUploadParams](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)
		return
	}

	_, err = checkFileAccessById(ctx, upInfo.RootFolderId)
	if err != nil {
		return
	}

	uploadEvent := journal.NewEvent()

	meta := job.UploadFilesMeta{
		ChunkStream:  make(chan job.FileChunk, 10),
		RootFolderId: upInfo.RootFolderId,
		ChunkSize:    upInfo.ChunkSize,
		User:         ctx.Requester, // The user who is starting the upload
		UploadEvent:  uploadEvent,
		Share:        ctx.Share,
	}

	t, err := ctx.TaskService.DispatchJob(ctx, job.UploadFilesTask, meta, nil)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)
		return
	}

	uploadInfo := structs.NewUploadInfo{UploadId: t.Id()}
	ctx.JSON(http.StatusCreated, uploadInfo)
}

// AddFilesToUpload godoc
//
//	@ID	AddFilesToUpload
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Add a file to an upload task
//	@Tags		Files
//	@Param		uploadId	path		string				true	"Upload Id"
//	@Param		shareId		query		string				false	"Share Id"
//	@Param		request		body		rest.NewFilesParams	true	"New file params"
//	@Success	201			{object}	rest.NewFilesInfo	"FileIds"
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/upload/{uploadId} [post]
func NewFileUpload(ctx *context.RequestContext) {
	uploadTaskId := ctx.Path("uploadId")
	params, err := net.ReadRequestBody[rest.NewFilesParams](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)
		return
	}

	uTask := ctx.TaskService.GetTask(uploadTaskId)
	if uTask == nil {
		ctx.Error(http.StatusNotFound, errors.New("No upload task exists with the given id"))
		return
	}

	completed, _ := uTask.Status()
	if completed {
		ctx.Error(http.StatusNotFound, errors.New("Upload task has already completed"))
		return
	}

	var ids []string
	for _, newFInfo := range params.NewFiles {
		parent, err := ctx.FileService.GetFileById(newFInfo.ParentFolderId)
		if err != nil {
			ctx.Error(http.StatusNotFound, errors.Wrap(err, "Could not find parent folder for new file"))
			return
		}

		child, _ := parent.GetChild(newFInfo.NewFileName)
		if child != nil {
			ctx.Error(http.StatusConflict, errors.New("File with the same name already exists in folder"))
			return
		}

		uTask.ClearTimeout()

		err = uTask.Manipulate(
			func(meta task_mod.TaskMetadata) error {
				uploadMeta := meta.(job.UploadFilesMeta)
				var newF *file_model.WeblensFileImpl
				if newFInfo.IsDir {
					newF, err = ctx.FileService.CreateFolder(parent, newFInfo.NewFileName)
					if err != nil {
						return err
					}
				} else {
					// We must not pass the event in here, as it will attempt to generate the contentId for the
					// file before the file has content.
					newF, err = ctx.FileService.CreateFile(parent, newFInfo.NewFileName, nil)
					if err != nil {
						return err
					}

					uploadMeta.ChunkStream <- job.FileChunk{
						NewFile: newF, ContentRange: "0-0/" + strconv.FormatInt(newFInfo.FileSize, 10),
					}
				}

				ids = append(ids, newF.ID())

				return nil
			},
		)

		if err != nil {
			ctx.Error(http.StatusInternalServerError, errors.Wrap(err, "Failed to add new file to upload task"))
			return
		}
	}

	newInfo := structs.NewFilesInfo{FileIds: ids}
	ctx.JSON(http.StatusCreated, newInfo)
}

// UploadFileChunk godoc
//
//	@ID	UploadFileChunk
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Add a chunk to a file upload
//	@Tags		Files
//	@Param		uploadId	path		string	true	"Upload Id"
//	@Param		fileId		path		string	true	"File Id"
//	@Param		shareId		query		string	false	"Share Id"
//	@Param		chunk		formData	file	true	"File chunk"
//	@Success	200
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/upload/{uploadId}/file/{fileId} [put]
func HandleUploadChunk(ctx *context.RequestContext) {
	uploadId := ctx.Path("uploadId")

	t := ctx.TaskService.GetTask(uploadId)
	if t == nil {
		ctx.Error(http.StatusNotFound, errors.New("No upload exists with given id"))
		return
	}

	fileId := ctx.Path("fileId")

	// We are about to read from the clientConn, which could take a while.
	// Since we actually got this request, we know the clientConn is not abandoning us,
	// so we can safely clear the timeout, which the task will re-enable if needed.
	t.ClearTimeout()

	chunk := make([]byte, ctx.Req.ContentLength)
	_, err := io.ReadAtLeast(ctx.Req.Body, chunk, int(ctx.Req.ContentLength))
	if err != nil {
		ctx.Logger.Error().Stack().Err(err).Msg("")
		// err = t.AddChunkToStream(fileId, nil, "0-0/-1")
		// if err != nil {
		// 	util.Error().Stack().Err(err).Msg("")
		// }
		ctx.Error(http.StatusInternalServerError, errors.New("Failed to read request body"))
		return
	}

	ctHeader := ctx.Header("Content-Type")
	if strings.HasPrefix(ctHeader, "multipart/form-data") {
		boundry := strings.Split(ctHeader, "boundary=")[1]
		chunk = bytes.TrimPrefix(chunk, []byte("--"+boundry+"\r\n"))
		chunk = bytes.TrimSuffix(chunk, []byte("--"+boundry+"--\r\n"))

		chunk = chunk[bytes.Index(chunk, []byte("\r\n\r\n"))+4:]
	}

	_, _, _, err = ctx.ContentRange()

	if err != nil {
		ctx.Error(http.StatusRequestedRangeNotSatisfiable, err)
		return
	}

	err = t.Manipulate(
		func(meta task_mod.TaskMetadata) error {
			chunkData := job.FileChunk{FileId: fileId, Chunk: chunk, ContentRange: ctx.Header("Content-Range")}
			meta.(job.UploadFilesMeta).ChunkStream <- chunkData

			return nil
		},
	)

	if err != nil {
		ctx.Error(http.StatusInternalServerError, errors.Wrap(err, "Failed to add chunk to upload task"))
		return
	}

	ctx.W.WriteHeader(http.StatusOK)
}

// GetUploadResult godoc
//
//	@ID	GetUploadResult
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Get the result of an upload task. This will block until the upload is complete
//	@Tags		Files
//	@Param		uploadId	path	string	true	"Upload Id"
//	@Success	200
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/upload/{uploadId} [get]
func GetUploadResult(ctx *context.RequestContext) {
	uploadId := ctx.Path("uploadId")
	if uploadId == "" {
		ctx.Error(http.StatusBadRequest, errors.New("Missing uploadId path parameter"))
		return
	}

	t := ctx.TaskService.GetTask(uploadId)
	if t == nil {
		ctx.Error(http.StatusNotFound, errors.New("No upload task exists with the given id"))
		return
	}

	t.Wait()
	ctx.W.WriteHeader(http.StatusOK)
}

// Helper Function
func getChildMedias(ctx *context.RequestContext, children []*file_model.WeblensFileImpl) ([]*media_model.Media, error) {
	var medias []*media_model.Media
	for _, child := range children {
		var m *media_model.Media
		contentId := child.GetContentId()
		if child.IsDir() && contentId == "" {
			coverId, err := ctx.FileService.GetFolderCover(child)
			if err != nil {
				return nil, err
			}

			if coverId != "" {
				child.SetContentId(coverId)
				contentId = coverId
			}
		}

		m, err := media_model.GetMediaById(ctx, contentId)
		if err != nil {
			ctx.Logger.Error().Stack().Err(err).Msgf("Failed to get media for file [%s]", child.ID())
			continue
		}

		medias = append(medias, m)
	}

	return medias, nil
}

// Helper Function
// Format and write back directory information. Authorization checks should be done before this function
func formatRespondFolderInfo(ctx *context.RequestContext, dir *file_model.WeblensFileImpl) error {

	var parentsInfo []structs.FileInfo
	parent := dir.GetParent()

	owner, err := ctx.FileService.GetFileOwner(parent)
	if err != nil {
		return err
	}

	auth.CanUserAccessFile(ctx.Requester, parent, ctx.Share) // Check if the user has access to the parent folder
	for parent.ID() != "ROOT" && auth.CanUserAccessFile(ctx.Requester, parent, ctx.Share) && !owner.IsSystemUser() {
		parentInfo, err := reshape.WeblensFileToFileInfo(ctx, parent, false)
		if err != nil {
			return err
		}
		parentsInfo = append(parentsInfo, parentInfo)
		parent = parent.GetParent()
	}

	children := dir.GetChildren()

	mediaFiles := append(children, dir)
	medias, err := getChildMedias(ctx, mediaFiles)
	if err != nil {
		return err
	}

	childInfos := make([]structs.FileInfo, 0, len(children))
	for _, child := range children {
		info, err := reshape.WeblensFileToFileInfo(ctx, child, false)
		if err != nil {
			return err
		}
		childInfos = append(childInfos, info)
	}

	selfInfo, err := reshape.WeblensFileToFileInfo(ctx, dir, false)
	if err != nil {
		return err
	}

	var mediaInfos []structs.MediaInfo
	for _, m := range medias {
		mediaInfos = append(mediaInfos, reshape.MediaToMediaInfo(m))
	}

	packagedInfo := structs.FolderInfoResponse{Self: selfInfo, Children: childInfos, Parents: parentsInfo, Medias: mediaInfos}
	ctx.JSON(http.StatusOK, packagedInfo)
	return nil
}

// Helper Function
func formatRespondPastFolderInfo(ctx *context.RequestContext, folderId string, pastTime time.Time) error {
	ctx.Logger.Trace().Func(func(e *zerolog.Event) { e.Msgf("Getting past folder [%s] at time [%s]", folderId, pastTime) })

	pastFile, err := journal.GetPastFile(folderId, pastTime)
	if err != nil {
		return err
	}
	pastFileInfo, err := reshape.WeblensFileToFileInfo(ctx, pastFile, true)
	if err != nil {
		return err
	}

	var parentsInfo []structs.FileInfo
	parentId := pastFile.GetParentId()
	if parentId == "" {
		return err
	}
	for parentId != "ROOT" {
		pastParent, err := journal.GetPastFile(parentId, pastTime)
		if err != nil {
			return err
		}

		parentInfo, err := reshape.WeblensFileToFileInfo(ctx, pastParent, true)
		if err != nil {
			return err
		}

		parentsInfo = append(parentsInfo, parentInfo)
		parentId = pastParent.GetParentId()
	}

	children := pastFile.GetChildren()
	childrenInfos := make([]structs.FileInfo, 0, len(children))
	for _, child := range children {
		if child.GetPortablePath().Filename() == file.UserTrashDirName {
			continue
		}

		childInfo, err := reshape.WeblensFileToFileInfo(ctx, child, true)
		if err != nil {
			return err
		}

		childrenInfos = append(childrenInfos, childInfo)
	}

	var medias []*media_model.Media
	for _, child := range children {
		m, err := media_model.GetMediaById(ctx, child.GetContentId())

		if err != nil {
			medias = append(medias, m)
		}
	}

	var mediaInfos []structs.MediaInfo
	for _, m := range medias {
		mediaInfos = append(mediaInfos, reshape.MediaToMediaInfo(m))
	}

	packagedInfo := structs.FolderInfoResponse{Self: pastFileInfo, Children: childrenInfos, Parents: parentsInfo, Medias: mediaInfos}
	ctx.JSON(http.StatusOK, packagedInfo)

	return nil
}
