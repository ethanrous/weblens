package file

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	cover_model "github.com/ethanrous/weblens/models/cover"
	"github.com/ethanrous/weblens/models/db"
	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/history"
	"github.com/ethanrous/weblens/models/job"
	media_model "github.com/ethanrous/weblens/models/media"
	share_model "github.com/ethanrous/weblens/models/share"
	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/netwrk"
	"github.com/ethanrous/weblens/modules/structs"
	"github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/services/auth"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	file_service "github.com/ethanrous/weblens/services/file"
	media_service "github.com/ethanrous/weblens/services/media"
	"github.com/ethanrous/weblens/services/notify"
	"github.com/ethanrous/weblens/services/reshape"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/rs/zerolog"
)

// GetFile godoc
//
//	@ID	GetFile
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Get information about a file
//	@Tags		Files
//	@Produce	json
//	@Param		fileID	path		string				true	"File ID"
//	@Param		shareID	query		string				false	"Share ID"
//	@Success	200		{object}	structs.FileInfo	"File Info"
//	@Failure	401
//	@Failure	404
//	@Router		/files/{fileID} [get]
func GetFile(ctx context_service.RequestContext) {
	file, err := checkFileAccess(ctx)
	if err != nil {
		return
	}

	fileInfo, err := reshape.WeblensFileToFileInfo(&ctx.AppContext, file)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

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
//	@Param		fileID	path		string	true	"File ID"
//	@Param		shareID	query		string	false	"Share ID"
//	@Success	200		{string}	string	"File text"
//	@Failure	400
//	@Router		/files/{fileID}/text [get]
func GetFileText(ctx context_service.RequestContext) {
	file, err := checkFileAccess(ctx)
	if err != nil {
		return
	}

	filename := file.GetPortablePath().Filename()

	dotIndex := strings.LastIndex(filename, ".")
	if filename[dotIndex:] != ".txt" {
		ctx.Error(http.StatusBadRequest, wlerrors.New("file is not a text file"))

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
//	@Param		fileID	path	string	true	"File ID"
//	@Failure	400
//	@Failure	501
//	@Router		/files/{fileID}/stats [get]
func GetFileStats(ctx context_service.RequestContext) {
	_, err := checkFileAccess(ctx)
	if err != nil {
		return
	}

	ctx.Status(http.StatusNotImplemented)
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
//	@Param		fileID		path		string						true	"File ID"
//	@Param		shareID		query		string						false	"Share ID"
//	@Param		format		query		string						false	"File format conversion"
//	@Param		isTakeout	query		bool						false	"Is this a takeout file"	Enums(true, false)	default(false)
//	@Success	200			{string}	binary						"File content"
//	@Success	404			{object}	structs.WeblensErrorInfo	"Error Info"
//	@Router		/files/{fileID}/download [get]
func DownloadFile(ctx context_service.RequestContext) {
	file, err := checkFileAccess(ctx)
	if err != nil {
		return
	}

	// TODO: Make sure to check if the requester is another tower
	// i := getInstanceFromCtx(r)
	//
	// if i != nil {
	// 	file, err = ctx.FileService.GetFileSafe(fileID, pack.UserService.GetRootUser(), nil)
	// 	if SafeErrorAndExit(err, w, log) {
	// 		return
	// 	}
	// }

	ctx.Log().Debug().Msgf("Headers?: %v", ctx.Req.Header)

	acceptType := ctx.Query("format")
	ctx.Log().Debug().Msgf("Accept type: [%s]", acceptType)

	if acceptType != "" && acceptType != "image/webp" {
		mt := media_model.ParseMime(acceptType)
		if mt.Name == "" {
			ctx.Error(http.StatusBadRequest, wlerrors.New("invalid format"))

			return
		}

		m, err := media_model.GetMediaByContentID(ctx, file.GetContentID())
		if err != nil {
			if wlerrors.Is(err, media_model.ErrMediaNotFound) {
				ctx.Error(http.StatusNotFound, err)

				return
			}

			ctx.Error(http.StatusInternalServerError, err)

			return
		}

		convertedImg, err := media_service.GetConverted(ctx, m, mt)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)

			return
		}

		ctx.Bytes(http.StatusOK, convertedImg)

		return
	}

	ctx.Log().Debug().Func(func(e *zerolog.Event) { e.Msgf("Downloading file %s", file.GetPortablePath()) })

	filePath := file.GetPortablePath().ToAbsolute()
	http.ServeFile(ctx.W, ctx.Req, filePath)
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
//	@Param		folderID	path		string						true	"Folder ID"
//	@Param		shareID		query		string						false	"Share ID"
//	@Param		timestamp	query		int							false	"Past timestamp to view the folder at, in ms since epoch"
//	@Success	200			{object}	structs.FolderInfoResponse	"Folder Info"
//	@Router		/folder/{folderID} [get]
func GetFolder(ctx context_service.RequestContext) {
	folder, err := checkFileAccess(ctx)
	if err != nil {
		return
	}

	date := time.UnixMilli(0)

	milliStr := ctx.Query("timestamp")
	if milliStr != "" {
		millis, err := strconv.ParseInt(milliStr, 10, 64)
		if err != nil {
			ctx.Error(http.StatusBadRequest, wlerrors.New("invalid timestamp format"))

			return
		}

		date = time.UnixMilli(millis)
	}

	if date.Unix() != 0 {
		err := formatRespondPastFolderInfo(ctx, folder, date)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err)
		}

		return
	}

	err = formatRespondFolderInfo(ctx, folder)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)
	}
}

// GetFolderHistory godoc
//
//	@ID			GetFolderHistory
//
//	@Security	SessionAuth
//
//	@Summary	Get actions of a folder at a given time
//	@Tags		Folder
//	@Param		fileID		path	string					true	"File ID"
//	@Success	200			{array}	structs.FileActionInfo	"File actions"
//	@Failure	400
//	@Failure	500
//	@Router		/files/{fileID}/history [get]
func GetFolderHistory(ctx context_service.RequestContext) {
	file, err := checkFileAccess(ctx)
	if err != nil {
		return
	}

	actions, err := history.GetActionsAtPathAfter(ctx, file.GetPortablePath(), time.Time{}, false)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	actionInfos := make([]structs.FileActionInfo, 0, len(actions))
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
//	@Param		search			query	string				true	"Filename to search for"
//	@Param		baseFolderID	query	string				false	"The folder to search in, defaults to the user's home folder"
//	@Success	200				{array}	structs.FileInfo	"File Info"
//	@Failure	400
//	@Failure	401
//	@Failure	500
//	@Router		/files/search [get]
func SearchByFilename(ctx context_service.RequestContext) {
	filenameSearch := ctx.Query("search")
	if filenameSearch == "" {
		ctx.Error(http.StatusBadRequest, wlerrors.New("missing 'search' query parameter"))

		return
	}

	baseFolderID := ctx.Query("baseFolderID")
	if baseFolderID == "" {
		baseFolderID = ctx.Requester.HomeID
	}

	baseFolder, err := CheckFileAccessByID(ctx, baseFolderID)
	if err != nil {
		return
	}

	if !baseFolder.IsDir() {
		ctx.Error(http.StatusBadRequest, wlerrors.New("the provided base folder ID is not a directory"))

		return
	}

	fileIDs := []string{}
	filenames := []string{}

	_ = baseFolder.RecursiveMap(
		func(f *file_model.WeblensFileImpl) error {
			fileIDs = append(fileIDs, f.ID())
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

	fileInfos := make([]structs.FileInfo, 0, len(matches))

	for _, match := range matches {
		f, err := ctx.FileService.GetFileByID(ctx, fileIDs[match.OriginalIndex])
		if err != nil {
			ctx.Log().Error().Stack().Err(err).Msgf("Failed to get file by ID: %s", fileIDs[match.OriginalIndex])

			continue
		}

		if f.ID() == ctx.Requester.HomeID || f.ID() == ctx.Requester.TrashID {
			continue
		}

		fileInfo, err := reshape.WeblensFileToFileInfo(&ctx.AppContext, f)
		if err != nil {
			ctx.Log().Error().Stack().Err(err).Msgf("Failed to convert file to FileInfo for file ID: %s", f.ID())

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
//	@Param		request	body		structs.CreateFolderBody	true	"New folder body"
//	@Param		shareID	query		string						false	"Share ID"
//	@Success	200		{object}	structs.FileInfo			"File Info"
//	@Router		/folder [post]
func CreateFolder(ctx context_service.RequestContext) {
	body, err := netwrk.ReadRequestBody[structs.CreateFolderBody](ctx.Req)
	if err != nil {
		return
	}

	if body.NewFolderName == "" {
		ctx.Error(http.StatusBadRequest, wlerrors.New("Missing body parameter 'newFolderName'"))

		return
	}

	parentFolder, err := CheckFileAccessByID(ctx, body.ParentFolderID, share_model.SharePermissionEdit)
	if err != nil {
		ctx.Error(http.StatusForbidden, wlerrors.New("You do not have permission to create a folder in this location"))

		return
	}

	ctx.Log().Trace().Msgf("User [%s] DOES have permission to create a folder in [%s]", ctx.Doer().Username, parentFolder.GetPortablePath())

	// var children []*file_model.WeblensFileImpl
	// if len(body.Children) != 0 {
	// 	for _, fileID := range body.Children {
	// 		child, err := ctx.FileService.GetFileSafe(fileID, u, nil)
	// 		if err != nil {
	// 			log.Error().Stack().Err(err).Msg("")
	// 			ctx.Error(http.StatusBadRequest, errors.New(err.Error()))
	// 			return
	// 		}
	// 		children = append(children, child)
	// 	}
	// }

	newDir, err := ctx.FileService.CreateFolder(ctx, parentFolder, body.NewFolderName)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	// err = ctx.FileService.MoveFiles(children, newDir, "USERS", pack.Caster)
	// if SafeErrorAndExit(err, w, log) {
	// 	return
	// }

	newDirInfo, err := reshape.WeblensFileToFileInfo(&ctx.AppContext, newDir)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.JSON(http.StatusOK, newDirInfo)
}

// SetFolderCover godoc
//
//	@ID			SetFolderCover
//
//	@Security	SessionAuth
//
//	@Summary	Set the cover image of a folder
//	@Tags		Folder
//	@Param		folderID	path	string	true	"Folder ID"
//	@Param		mediaID		query	string	true	"Media ID"
//	@Success	200
//	@Failure	400
//	@Failure	404
//	@Failure	500
//	@Router		/folder/{folderID}/cover [patch]
func SetFolderCover(ctx context_service.RequestContext) {
	folder, err := checkFileAccess(ctx)
	if err != nil {
		return
	}

	mediaID := ctx.Query("mediaID")

	media, err := media_model.GetMediaByContentID(ctx, mediaID)
	if err != nil {
		if wlerrors.Is(err, media_model.ErrMediaNotFound) {
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

	fInfo, err := reshape.WeblensFileToFileInfo(ctx, folder)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	notif := notify.NewFileNotification(ctx, fInfo, websocket.FileUpdatedEvent)
	ctx.Notify(ctx, notif...)

	ctx.Status(http.StatusOK)
}

// GetSharedFiles godoc
//
//	@ID			GetSharedFiles
//
//	@Security	SessionAuth
//
//	@Summary	Get files shared with the logged in user
//	@Tags		Files
//	@Success	200	{object}	structs.FolderInfoResponse	"All the top-level files shared with the user"
//	@Failure	404
//	@Failure	500
//	@Router		/files/shared [get]
func GetSharedFiles(ctx context_service.RequestContext) {
	shares, err := share_model.GetSharedWithUser(ctx, ctx.Requester.Username)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	children := make([]*file_model.WeblensFileImpl, 0, len(shares))

	for _, share := range shares {
		f, err := ctx.FileService.GetFileByID(ctx, share.FileID)
		if err != nil {
			if wlerrors.Is(err, file_model.ErrFileNotFound) {
				ctx.Log().Error().Stack().Err(err).Msg("Could not find file acompanying a file share")

				continue
			}

			ctx.Error(http.StatusInternalServerError, err)

			return
		}

		children = append(children, f)
	}

	childInfos := make([]structs.FileInfo, 0, len(children))

	for _, child := range children {
		fInfo, err := reshape.WeblensFileToFileInfo(&ctx.AppContext, child)
		if err != nil {
			// If we can't convert the file to a FileInfo, log the error and continue
			ctx.Error(http.StatusInternalServerError, wlerrors.New("failed to convert file to FileInfo"))

			return
		}

		childInfos = append(childInfos, fInfo)
	}

	mediaInfos := make([]structs.MediaInfo, 0)

	for _, child := range children {
		var media *media_model.Media

		if child.IsDir() {
			cover, err := cover_model.GetCoverByFolderID(ctx, child.ID()) // This will return the cover media if it exists
			if db.IsNotFound(err) {
				continue
			} else if err != nil {
				// If there was an error retrieving the cover media, log the error and continue
				ctx.Log().Error().Stack().Err(err).Msg("Failed to get cover media for folder")

				continue
			}

			media, err = media_model.GetMediaByContentID(ctx, cover.CoverPhotoID) // This will ensure we have the media object to send back to the client
			if err != nil {
				ctx.Log().Error().Stack().Err(err).Msg("Failed to get media for cover photo")

				continue
			}
		} else {
			media, err = media_model.GetMediaByContentID(ctx, child.GetContentID()) // This will ensure we have the media object to send back to the client
			if err != nil {
				continue
			}
		}

		mediaInfo := reshape.MediaToMediaInfo(media) // This will convert the media object to a MediaInfo object for the response
		mediaInfos = append(mediaInfos, mediaInfo)
	}

	fakeSelfFile := structs.FileInfo{
		ID:           "shared",
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
//	@Param			shareID	query		string					false	"Share ID"
//	@Param			request	body		structs.FilesListParams	true	"File Ids"
//	@Success		200		{object}	structs.TakeoutInfo		"Zip Takeout Info"
//	@Success		202		{object}	structs.TakeoutInfo		"Task Dispatch Info"
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/takeout [post]
func CreateTakeout(ctx context_service.RequestContext) {
	takeoutRequest, err := netwrk.ReadRequestBody[structs.FilesListParams](ctx.Req)
	if err != nil {
		return
	}

	if len(takeoutRequest.FileIDs) == 0 {
		ctx.Error(http.StatusBadRequest, wlerrors.New("Cannot takeout 0 files"))

		return
	}

	// TODO: Make sure user has access to all requested files
	files := make([]*file_model.WeblensFileImpl, 0, len(takeoutRequest.FileIDs))

	for _, fileID := range takeoutRequest.FileIDs {
		file, err := ctx.FileService.GetFileByID(ctx, fileID)
		if err != nil {
			ctx.Error(http.StatusNotFound, err)

			return
		}

		files = append(files, file)
	}

	// If we only have 1 file, and it is not a directory, we should have requested to just
	// simply download that file on it's own, not zip it.
	if len(files) == 1 && !files[0].IsDir() {
		ctx.Error(http.StatusBadRequest, wlerrors.New("Single non-directory file should not be zipped"))

		return
	}

	meta := job.ZipMeta{
		Files:     files,
		Requester: ctx.Requester,
		Share:     ctx.Share,
	}

	t, err := ctx.TaskService.DispatchJob(ctx, job.CreateZipTask, meta, nil)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, wlerrors.Wrap(err, "Failed to dispatch zip task"))

		return
	}

	completed, status := t.Status()
	if completed && status == task.TaskSuccess {
		result := t.GetResult()
		res := structs.TakeoutInfo{TakeoutID: result["takeoutID"].(string), Single: false, Filename: result["filename"].(string)}
		ctx.JSON(http.StatusOK, res)
	} else {
		ctx.JSON(http.StatusAccepted, structs.TakeoutInfo{TaskID: t.ID()})
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
//	@Param		searchPath	query		string						true	"Search path"
//	@Success	200			{object}	structs.FolderInfoResponse	"Path info"
//	@Failure	500
//	@Router		/files/autocomplete [get]
func AutocompletePath(ctx context_service.RequestContext) {
	searchPath := ctx.Query("searchPath")
	if len(searchPath) == 0 {
		ctx.Error(http.StatusBadRequest, wlerrors.New("Missing 'searchPath' query parameter"))

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

	folder, err := ctx.FileService.GetFileByFilepath(ctx, filepath)
	if err != nil {
		ctx.Error(http.StatusNotFound, err)

		return
	}

	children := folder.GetChildren()
	if folder.GetParent().ID() == "ROOT" {
		trashIndex := slices.IndexFunc(children, func(f *file_model.WeblensFileImpl) bool {
			return f.ID() == ctx.Requester.TrashID
		})
		children = slices.Delete(children, trashIndex, trashIndex+1)
	}

	filenames := make([]string, 0, len(children))
	for _, child := range children {
		filenames = append(filenames, child.GetPortablePath().Filename())
	}

	matches := fuzzy.RankFindFold(childName, filenames)
	slices.SortFunc(
		matches, func(a, b fuzzy.Rank) int {
			diff := a.Distance - b.Distance
			if diff == 0 {
				return strings.Compare(a.Target, b.Target)
			}

			return diff
		},
	)

	childInfos := make([]structs.FileInfo, 0, len(matches))

	for _, match := range matches {
		f := children[match.OriginalIndex]

		childInfo, err := reshape.WeblensFileToFileInfo(ctx.AppContext, f)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, wlerrors.New("failed to convert file to FileInfo"))

			return
		}

		childInfos = append(childInfos, childInfo)
	}

	selfInfo, err := reshape.WeblensFileToFileInfo(ctx.AppContext, folder)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, wlerrors.New("failed to convert folder to FileInfo"))

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
//	@Summary	structsore files from some time in the past
//	@Tags		Files
//	@Accept		json
//	@Produce	json
//	@Param		request	body		structs.RestoreFilesParams	true	"RestoreFiles files request body"
//	@Success	200		{object}	structs.RestoreFilesInfo	"structsore files info"
//	@Failure	400
//	@Failure	404
//	@Failure	500
//	@Router		/files/structsore [post]
func RestoreFiles(ctx context_service.RequestContext) {
	body, err := netwrk.ReadRequestBody[structs.RestoreFilesParams](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, wlerrors.New("Failed to read request body"))
	}

	if body.Timestamp == 0 {
		ctx.Error(http.StatusBadRequest, wlerrors.New("Missing body parameter 'timestamp'"))

		return
	}

	ctx.Status(http.StatusNotImplemented)
	// structsoreTime := time.UnixMilli(body.Timestamp)

	// parentLt := ctx.FileService.GetJournalByTree("USERS").Get(body.NewParentID)
	// if parentLt == nil {
	// 	ctx.Error(http.StatusNotFound, errors.New("Could not find new parent"))
	// 	return
	// }
	//
	// // New parent folder is the folder it was in at the time we are structsoring from, if
	// // it still exists, otherwise it is the users home folder
	// var newParent *file_model.WeblensFileImpl
	// if parentLt.GetLatestAction().GetActionType() == fileTree.FileDelete {
	// 	newParent, err = ctx.FileService.GetFileSafe(u.HomeID, u, nil)
	//
	// 	// this should never error, but you never know
	// 	if SafeErrorAndExit(err, w, log) {
	// 		return
	// 	}
	// } else {
	// 	newParent, err = ctx.FileService.GetFileSafe(body.NewParentID, u, nil)
	// 	if SafeErrorAndExit(err, w, log) {
	// 		return
	// 	}
	// }
	//
	// // actions := parentLt.GetActions()
	// // for i, action := range actions {
	// // 	if action.Timestamp.After(structsoreTime) && (action.ActionType != fileTree.FileSizeChange || i == len(actions)-1) {
	// // 		if i != 0 {
	// // 			structsoreTime = actions[i-1].Timestamp
	// // 		}
	// // 		break
	// // 	}
	// // }
	//
	// err = ctx.FileService.structsoreFiles(body.FileIds, newParent, structsoreTime, pack.Caster)
	// if SafeErrorAndExit(err, w, log) {
	// 	return
	// }
	// res := structs.structsoreFilesInfo{NewParentID: newParent.ID()}
	// writeJSON(w, http.StatusOK, res)
	_ = ""
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
//	@Param		fileID	path	string						true	"File ID"
//	@Param		shareID	query	string						false	"Share ID"
//	@Param		request	body	structs.UpdateFileParams	true	"Update file request body"
//	@Success	200
//	@Failure	403
//	@Failure	404
//	@Failure	500
//	@Router		/files/{fileID} [patch]
func UpdateFile(ctx context_service.RequestContext) {
	file, err := checkFileAccess(ctx)
	if err != nil {
		return
	}

	if file_model.IsFileInTrash(file) {
		ctx.Error(http.StatusForbidden, wlerrors.New("cannot rename file in trash"))

		return
	}

	updateInfo, err := netwrk.ReadRequestBody[structs.UpdateFileParams](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, wlerrors.New("Failed to read request body"))

		return
	}

	err = ctx.FileService.RenameFile(ctx, file, updateInfo.NewName)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.Status(http.StatusOK)
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
//	@Param		request	body	structs.MoveFilesParams	true	"Move files request body"
//	@Param		shareID	query	string					false	"Share ID"
//	@Success	200
//	@Failure	404
//	@Failure	500
//	@Router		/files [patch]
func MoveFiles(ctx context_service.RequestContext) {
	filesData, err := netwrk.ReadRequestBody[structs.MoveFilesParams](ctx.Req)
	if err != nil {
		return
	}

	newParent, err := CheckFileAccessByID(ctx, filesData.NewParentID, share_model.SharePermissionEdit)
	if err != nil {
		return
	}

	if newParent.IsDir() == false {
		ctx.Error(http.StatusBadRequest, wlerrors.New("New parent is not a directory"))

		return
	}

	if file_model.IsFileInTrash(newParent) {
		if _, err = auth.CanUserAccessFile(ctx, ctx.Requester, newParent, ctx.Share, share_model.SharePermissionDelete); err != nil {
			// If the user does not have permission to delete, return forbidden
			ctx.Error(http.StatusForbidden, err)

			return
		}
	}

	if len(filesData.Files) == 0 {
		ctx.Error(http.StatusBadRequest, wlerrors.New("No file ids provided"))

		return
	}

	perms := []share_model.Permission{share_model.SharePermissionEdit}
	// if the new parent is in the trash, we need to check delete permission as well
	if file_model.IsFileInTrash(newParent) {
		perms = append(perms, share_model.SharePermissionDelete)
	}

	files := make([]*file_model.WeblensFileImpl, 0, len(filesData.Files))

	for _, fileID := range filesData.Files {
		f, err := ctx.FileService.GetFileByID(ctx, fileID)
		if err != nil {
			ctx.Error(http.StatusNotFound, wlerrors.New("Could not find file with id "+fileID))

			return
		}

		if _, err = auth.CanUserAccessFile(ctx, ctx.Requester, f, ctx.Share, perms...); err != nil {
			// If the user does not have permission to delete, return forbidden
			ctx.Error(http.StatusForbidden, err)

			return
		}

		files = append(files, f)
	}

	err = db.WithTransaction(ctx, func(sessCtx context.Context) error {
		return ctx.FileService.MoveFiles(sessCtx, files, newParent)
	})
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	ctx.Status(http.StatusOK)
}

// UnTrashFiles godoc
//
//	@ID			UnTrashFiles
//
//	@Security	SessionAuth
//
//	@Summary	Move a list of files out of the trash, structsoring them to where they were before
//	@Tags		Files
//	@Param		request	body	structs.FilesListParams	true	"Un-trash files request body"
//	@Success	200
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/files/untrash [patch]
func UnTrashFiles(ctx context_service.RequestContext) {
	params, err := netwrk.ReadRequestBody[structs.FilesListParams](ctx.Req)
	if err != nil {
		return
	}

	fileIDs := params.FileIDs
	files := make([]*file_model.WeblensFileImpl, 0, len(fileIDs))

	for _, fileID := range fileIDs {
		file, err := ctx.FileService.GetFileByID(ctx, fileID)
		if err != nil {
			ctx.Error(http.StatusNotFound, wlerrors.New("Could not find file with id "+fileID))

			return
		}

		files = append(files, file)
	}

	err = ctx.FileService.ReturnFilesFromTrash(ctx, files)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, wlerrors.Wrap(err, "Failed to un-trash files"))
	}

	ctx.Status(http.StatusOK)
}

// DeleteFiles godoc
//
//	@ID			DeleteFiles
//
//	@Security	SessionAuth
//
//	@Summary	Delete Files "permanently"
//	@Tags		Files
//	@Param		request			body	structs.FilesListParams	true	"Delete files request body"
//	@Param		ignoreTrash		query	boolean					false	"Delete files even if they are not in the trash"
//	@Param		preserveFolder	query	boolean					false	"Preserve parent folder if it is empty after deletion"
//	@Success	200
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/files [delete]
func DeleteFiles(ctx context_service.RequestContext) {
	params, err := netwrk.ReadRequestBody[structs.FilesListParams](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	if len(params.FileIDs) == 0 {
		ctx.Error(http.StatusBadRequest, wlerrors.New("No file ids provided"))

		return
	}

	files := make([]*file_model.WeblensFileImpl, 0, len(params.FileIDs))

	for _, fileID := range params.FileIDs {
		file, err := CheckFileAccessByID(ctx, fileID, share_model.SharePermissionDelete)
		if err != nil {
			ctx.Error(http.StatusForbidden, err)

			return
		}

		files = append(files, file)
	}

	preserveFolder := ctx.QueryBool("preserveFolder")
	ctx.Log().Debug().Msgf("Preserve folder: %v", preserveFolder)

	if preserveFolder {
		childFiles := []*file_model.WeblensFileImpl{}

		for _, file := range files {
			if file_model.IsFileInTrash(file) && file.GetPortablePath().Filename() != file_model.UserTrashDirName {
				// If the file is in the trash, we can delete it without preserving the folder
				childFiles = append(childFiles, file)
			} else {
				// Otherwise get the children of the file and delete those instead
				children, err := ctx.FileService.GetChildren(ctx, file)
				if err != nil {
					ctx.Error(http.StatusInternalServerError, fmt.Errorf("Failed to get children of file: %w", err))

					return
				}

				childFiles = append(childFiles, children...)
			}
		}

		files = childFiles
	}

	err = ctx.FileService.DeleteFiles(ctx, files...)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, fmt.Errorf("Failed to delete files: %w", err))

		return
	}

	ctx.Status(http.StatusOK)
}

const chunkChanSize = 10

// NewUploadTask godoc
//
//	@ID	StartUpload
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Begin a new upload task
//	@Tags		Files
//	@Param		request	body		structs.NewUploadParams	true	"New upload request body"
//	@Param		shareID	query		string					false	"Share ID"
//	@Success	201		{object}	structs.NewUploadInfo	"Upload Info"
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/upload [post]
func NewUploadTask(ctx context_service.RequestContext) {
	upInfo, err := netwrk.ReadRequestBody[structs.NewUploadParams](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	_, err = CheckFileAccessByID(ctx, upInfo.RootFolderID, share_model.SharePermissionEdit)
	if err != nil {
		return
	}

	meta := job.UploadFilesMeta{
		ChunkStream:  make(chan job.FileChunk, chunkChanSize),
		RootFolderID: upInfo.RootFolderID,
		ChunkSize:    upInfo.ChunkSize,
		User:         ctx.Requester, // The user who is starting the upload
		Share:        ctx.Share,
	}

	t, err := ctx.TaskService.DispatchJob(ctx, job.UploadFilesTask, meta, nil)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	uploadInfo := structs.NewUploadInfo{UploadID: t.ID()}
	ctx.JSON(http.StatusCreated, uploadInfo)
}

// NewFileUpload godoc
//
//	@ID	AddFilesToUpload
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Add a file to an upload task
//	@Tags		Files
//	@Param		uploadID	path		string					true	"Upload ID"
//	@Param		shareID		query		string					false	"Share ID"
//	@Param		request		body		structs.NewFilesParams	true	"New file params"
//	@Success	201			{object}	structs.NewFilesInfo	"FileIds"
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/upload/{uploadID} [post]
func NewFileUpload(ctx context_service.RequestContext) {
	uploadTaskID := ctx.Path("uploadID")

	params, err := netwrk.ReadRequestBody[structs.NewFilesParams](ctx.Req)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, err)

		return
	}

	uTask := ctx.TaskService.GetTask(uploadTaskID)
	if uTask == nil {
		ctx.Error(http.StatusNotFound, wlerrors.New("No upload task exists with the given id"))

		return
	}

	completed, _ := uTask.Status()
	if completed {
		ctx.Error(http.StatusNotFound, wlerrors.New("Upload task has already completed"))

		return
	}

	var ids []string

	for _, newFInfo := range params.NewFiles {
		parent, err := ctx.FileService.GetFileByID(ctx, newFInfo.ParentFolderID)
		if err != nil {
			ctx.Error(http.StatusNotFound, wlerrors.Wrap(err, "Could not find parent folder for new file"))

			return
		}

		child, _ := parent.GetChild(newFInfo.NewFileName)
		if child != nil {
			ctx.Error(http.StatusConflict, wlerrors.New("File with the same name already exists in folder"))

			return
		}

		uTask.ClearTimeout()

		err = uTask.Manipulate(
			func(meta task.Metadata) error {
				uploadMeta := meta.(job.UploadFilesMeta)

				var newF *file_model.WeblensFileImpl

				if newFInfo.IsDir {
					newF, err = ctx.FileService.CreateFolder(ctx, parent, newFInfo.NewFileName)
					if err != nil {
						return err
					}
				} else {
					// We must not pass the event in here, as it will attempt to generate the contentID for the
					// file before the file has content.
					newF, err = ctx.FileService.CreateFile(ctx.WithValue(file_service.SkipJournalKey, true), parent, newFInfo.NewFileName)
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
			ctx.Error(http.StatusInternalServerError, wlerrors.Wrap(err, "Failed to add new file to upload task"))

			return
		}
	}

	newInfo := structs.NewFilesInfo{FileIDs: ids}
	ctx.JSON(http.StatusCreated, newInfo)
}

// HandleUploadChunk godoc
//
//	@ID	UploadFileChunk
//
//	@Security
//	@Security	SessionAuth
//
//	@Summary	Add a chunk to a file upload
//	@Tags		Files
//	@Param		uploadID	path		string	true	"Upload ID"
//	@Param		fileID		path		string	true	"File ID"
//	@Param		shareID		query		string	false	"Share ID"
//	@Param		chunk		formData	file	true	"File chunk"
//	@Success	200
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/upload/{uploadID}/file/{fileID} [put]
func HandleUploadChunk(ctx context_service.RequestContext) {
	uploadID := ctx.Path("uploadID")

	t := ctx.TaskService.GetTask(uploadID)
	if t == nil {
		ctx.Error(http.StatusNotFound, wlerrors.New("No upload exists with given id"))

		return
	}

	// if rand.IntN(100) >= 95 {
	// 	ctx.Error(http.StatusBadRequest, errors.New("Fake error"))
	//
	// 	return
	// }
	//

	fileID := ctx.Path("fileID")

	// We are about to read from the clientConn, which could take a while.
	// Since we actually got this request, we know the clientConn is not abandoning us,
	// so we can safely clear the timeout, which the task will re-enable if needed.
	t.ClearTimeout()

	chunk := make([]byte, ctx.Req.ContentLength)

	readBs, err := io.ReadAtLeast(ctx.Req.Body, chunk, int(ctx.Req.ContentLength))
	if err != nil {
		err = fmt.Errorf("expected to read exactly %d bytes, but read %d: %w", ctx.Req.ContentLength, readBs, err)
		ctx.Log().Error().Stack().Err(err).Msgf("")
		ctx.Error(http.StatusBadRequest, err)

		return
	}

	ctHeader := ctx.Header("Content-Type")
	if strings.HasPrefix(ctHeader, "multipart/form-data") {
		boundary := strings.Split(ctHeader, "boundary=")[1]
		chunk = bytes.TrimPrefix(chunk, []byte("--"+boundary+"\r\n"))
		chunk = bytes.TrimSuffix(chunk, []byte("--"+boundary+"--\r\n"))

		chunk = chunk[bytes.Index(chunk, []byte("\r\n\r\n"))+4:]
	}

	_, _, _, err = ctx.ContentRange()
	if err != nil {
		ctx.Error(http.StatusRequestedRangeNotSatisfiable, err)

		return
	}

	err = t.Manipulate(
		func(meta task.Metadata) error {
			chunkData := job.FileChunk{FileID: fileID, Chunk: chunk, ContentRange: ctx.Header("Content-Range")}
			meta.(job.UploadFilesMeta).ChunkStream <- chunkData

			return nil
		},
	)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, wlerrors.Wrap(err, "Failed to add chunk to upload task"))

		return
	}

	ctx.Status(http.StatusOK)
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
//	@Param		uploadID	path	string	true	"Upload ID"
//	@Success	200
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/upload/{uploadID} [get]
func GetUploadResult(ctx context_service.RequestContext) {
	uploadID := ctx.Path("uploadID")
	if uploadID == "" {
		ctx.Error(http.StatusBadRequest, wlerrors.New("Missing uploadID path parameter"))

		return
	}

	t := ctx.TaskService.GetTask(uploadID)
	if t == nil {
		ctx.Error(http.StatusNotFound, wlerrors.New("No upload task exists with the given id"))

		return
	}

	t.Wait()
	ctx.Status(http.StatusOK)
}

// Helper Function.
func getChildMedias(
	ctx context_service.RequestContext,
	children []*file_model.WeblensFileImpl,
) ([]*media_model.Media, error) {
	medias := []*media_model.Media{}

	for _, child := range children {
		if child.IsDir() && child.GetContentID() == "" {
			cover, err := cover_model.GetCoverByFolderID(ctx, child.ID())

			if db.IsNotFound(err) {
				// No cover for this folder, skip it
				continue
			} else if err != nil {
				ctx.Log().Error().Stack().Err(err).Msgf("failed to get child media")

				continue
			}

			child.SetContentID(cover.CoverPhotoID)
		}

		if child.GetContentID() == "" {
			continue
		}

		var m *media_model.Media

		m, err := media_model.GetMediaByContentID(ctx, child.GetContentID())
		if err != nil && !db.IsNotFound(err) {
			return nil, err
		} else if err != nil {
			continue
		}

		medias = append(medias, m)
	}

	return medias, nil
}
