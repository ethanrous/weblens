package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/rest"
	"github.com/ethanrous/weblens/task"
	"github.com/go-chi/chi/v5"
	"github.com/lithammer/fuzzysearch/fuzzy"
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
//	@Param		fileId	path		string			true	"File Id"
//	@Param		shareId	query		string			false	"Share Id"
//	@Success	200		{object}	rest.FileInfo	"File Info"
//	@Failure	401
//	@Failure	404
//	@Router		/files/{fileId} [get]
func getFile(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}
	if u == nil {
		return
	}
	sh, err := getShareFromCtx[*models.FileShare](w, r)
	if err != nil {
		return
	}

	fileId := chi.URLParam(r, "fileId")
	file, err := pack.FileService.GetFileSafe(fileId, u, sh)
	if err != nil {
		log.ShowErr(err)
		writeJson(w, http.StatusNotFound, rest.WeblensErrorInfo{Error: "Could not find file"})
		return
	}

	if file == nil {
		writeJson(w, http.StatusNotFound, rest.WeblensErrorInfo{Error: "Could not find file"})
		return
	}

	formattedInfo, err := rest.WeblensFileToFileInfo(file, pack, false)
	if SafeErrorAndExit(err, w) {
		return
	}

	writeJson(w, http.StatusOK, formattedInfo)
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
//	@Success	200		{string}	string	"File text"
//	@Failure	400
//	@Router		/files/{fileId}/text [get]
func getFileText(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}
	if u == nil {
		return
	}

	fileId := chi.URLParam(r, "fileId")
	file, err := pack.FileService.GetFileSafe(fileId, u, nil)
	if SafeErrorAndExit(err, w) {
		return
	}

	filename := file.Filename()

	dotIndex := strings.LastIndex(filename, ".")
	if filename[dotIndex:] != ".txt" {
		writeJson(w, http.StatusBadRequest, rest.WeblensErrorInfo{Error: "File is not a text file"})
		return
	}

	http.ServeFile(w, r, file.Filename())
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
func getFileStats(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}
	if u == nil {
		u = pack.UserService.GetPublicUser()
	}

	fileId := chi.URLParam(r, "fileId")

	rootFolder, err := pack.FileService.GetFileSafe(fileId, u, nil)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if !rootFolder.IsDir() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNotImplemented)

	// t := types.SERV.TaskDispatcher.GatherFsStats(rootFolder, Caster)
	// t.Wait()
	// res := t.GetResult("sizesByExtension")
	//
	// writeJson(w, comm.StatusOK, res)
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
//	@Param		isTakeout	query		bool					false	"Is this a takeout file"	Enums(true, false)	default(false)
//	@Success	200			{string}	binary					"File content"
//	@Success	404			{object}	rest.WeblensErrorInfo	"Error Info"
//	@Router		/files/{fileId}/download [get]
func downloadFile(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}

	fileId := chi.URLParam(r, "fileId")
	isTakeout := r.URL.Query().Get("isTakeout")

	var file *fileTree.WeblensFileImpl
	if isTakeout == "true" {
		var err error
		file, err = pack.FileService.GetZip(fileId)
		if err != nil {
			writeJson(w, http.StatusNotFound, err)
			return
		}
	} else {
		share, err := getShareFromCtx[*models.FileShare](w, r)
		if err != nil {
			return
		}

		file, err = pack.FileService.GetFileSafe(fileId, u, share)
		if err != nil {
			writeJson(w, http.StatusNotFound, err)
			return
		}
	}

	http.ServeFile(w, r, file.AbsPath())
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
func getFolderHistory(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)

	fileId := chi.URLParam(r, "fileId")
	if fileId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pastTime := time.Now()

	milliStr := r.URL.Query().Get("timestamp")
	if milliStr != "" {
		millis, err := strconv.ParseInt(milliStr, 10, 64)
		if err != nil {
			log.ShowErr(werror.WithStack(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		pastTime = time.UnixMilli(millis)
	}

	f, err := pack.FileService.GetJournalByTree("USERS").GetPastFile(fileId, pastTime)
	if SafeErrorAndExit(err, w) {
		return
	}

	actions, err := pack.FileService.GetJournalByTree("USERS").GetActionsByPath(f.GetPortablePath())
	if SafeErrorAndExit(err, w) {
		return
	}

	var actionInfos []rest.FileActionInfo
	for _, a := range actions {
		actionInfos = append(actionInfos, rest.FileActionToFileActionInfo(a))
	}

	writeJson(w, http.StatusOK, actionInfos)
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
func searchByFilename(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}

	filenameSearch := r.URL.Query().Get("search")
	if filenameSearch == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	baseFolderId := r.URL.Query().Get("baseFolderId")
	if baseFolderId == "" {
		baseFolderId = u.HomeId
	}

	baseFolder, err := pack.FileService.GetFileSafe(baseFolderId, u, nil)
	if SafeErrorAndExit(err, w) {
		return
	}

	if !baseFolder.IsDir() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var fileIds []fileTree.FileId
	var filenames []string
	_ = baseFolder.RecursiveMap(
		func(f *fileTree.WeblensFileImpl) error {
			fileIds = append(fileIds, f.ID())
			filenames = append(filenames, f.Filename())
			return nil
		},
	)

	matches := fuzzy.RankFindFold(filenameSearch, filenames)
	slices.SortFunc(
		matches, func(a, b fuzzy.Rank) int {
			return a.Distance - b.Distance
		},
	)

	files := internal.FilterMap(
		matches, func(match fuzzy.Rank) (*fileTree.WeblensFileImpl, bool) {
			f, err := pack.FileService.GetFileSafe(fileIds[match.OriginalIndex], u, nil)
			if err != nil {
				return nil, false
			}
			if f.ID() == u.HomeId || f.ID() == u.TrashId {
				return nil, false
			}
			return f, true
		},
	)

	var fileInfos []rest.FileInfo
	for _, file := range files {
		f, err := rest.WeblensFileToFileInfo(file, pack, false)
		if SafeErrorAndExit(err, w) {
			return
		}
		fileInfos = append(fileInfos, f)
	}

	writeJson(w, http.StatusOK, fileInfos)
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
func createFolder(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}

	body, err := readCtxBody[rest.CreateFolderBody](w, r)
	if err != nil {
		return
	}

	if body.NewFolderName == "" {
		writeJson(w, http.StatusBadRequest, rest.WeblensErrorInfo{Error: "Missing body parameter 'newFolderName'"})
		return
	}

	parentFolder, err := pack.FileService.GetFileSafe(body.ParentFolderId, u, nil)
	if err != nil {
		writeJson(w, http.StatusNotFound, rest.WeblensErrorInfo{Error: "Parent folder not found"})
		return
	}

	var children []*fileTree.WeblensFileImpl
	if len(body.Children) != 0 {
		for _, fileId := range body.Children {
			child, err := pack.FileService.GetFileSafe(fileId, u, nil)
			if err != nil {
				log.ShowErr(err)
				writeJson(w, http.StatusBadRequest, rest.WeblensErrorInfo{Error: err.Error()})
				return
			}
			children = append(children, child)
		}
	}

	newDir, err := pack.FileService.CreateFolder(parentFolder, body.NewFolderName, nil, pack.Caster)
	if SafeErrorAndExit(err, w) {
		return
	}

	err = pack.FileService.MoveFiles(children, newDir, "USERS", pack.Caster)
	if SafeErrorAndExit(err, w) {
		return
	}

	newDirInfo, err := rest.WeblensFileToFileInfo(newDir, pack, false)
	if SafeErrorAndExit(err, w) {
		return
	}

	writeJson(w, http.StatusCreated, newDirInfo)
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
//	@Param		folderId	path		string	true	"Folder Id"
//	@Param		shareId		query		string	false	"Share Id"
//	@Param		timestamp	query		int		false	"Past timestamp to view the folder at, in ms since epoch"
//	@Success	200			{object}	rest.FolderInfoResponse"Folder Info"
//	@Router		/folder/{folderId} [get]
func getFolder(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}
	sh, err := getShareFromCtx[*models.FileShare](w, r)
	if err != nil {
		return
	}

	milliStr := r.URL.Query().Get("timestamp")
	date := time.UnixMilli(0)
	if milliStr != "" {
		millis, err := strconv.ParseInt(milliStr, 10, 64)
		if SafeErrorAndExit(err, w) {
			return
		} else if millis < 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		date = time.UnixMilli(millis)
	}

	folderId := chi.URLParam(r, "folderId")
	if folderId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if date.Unix() != 0 {
		formatRespondPastFolderInfo(folderId, date, w, r)
		return
	}

	dir, err := pack.FileService.GetFileSafe(folderId, u, sh)
	if SafeErrorAndExit(err, w) {
		return
	}

	formatRespondFolderInfo(dir, w, r)
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
func setFolderCover(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}

	folderId := chi.URLParam(r, "folderId")
	mediaId := r.URL.Query().Get("mediaId")

	if folderId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	folder, err := pack.FileService.GetFileSafe(folderId, u, nil)
	if SafeErrorAndExit(err, w) {
		return
	}

	coverId, err := pack.FileService.GetFolderCover(folder)
	if SafeErrorAndExit(err, w) {
		return
	}

	if coverId == "" && mediaId == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if coverId == mediaId {
		w.WriteHeader(http.StatusOK)
		return
	}

	var m *models.Media
	if mediaId != "" {
		m = pack.MediaService.Get(mediaId)
		if m == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	err = pack.FileService.SetFolderCover(folderId, mediaId)
	if SafeErrorAndExit(err, w) {
		return
	}

	pack.Caster.PushFileUpdate(folder, m)

	w.WriteHeader(http.StatusOK)
}

func getExternalDirs(w http.ResponseWriter, r *http.Request) {
	panic(werror.NotImplemented("getExternalDirs"))
}

func getExternalFolderInfo(w http.ResponseWriter, r *http.Request) {
	panic(werror.NotImplemented("getExternalFolderInfo"))
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
func scanDir(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}
	sh, err := getShareFromCtx[*models.FileShare](w, r)
	if err != nil {
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var scanInfo rest.ScanBody
	err = json.Unmarshal(body, &scanInfo)
	if err != nil {
		log.ErrTrace(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dir, err := pack.FileService.GetFileSafe(scanInfo.FolderId, u, sh)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	caster := models.NewSimpleCaster(pack.ClientService)
	meta := models.ScanMeta{
		File:         dir,
		FileService:  pack.FileService,
		MediaService: pack.MediaService,
		TaskSubber:   pack.ClientService,
		TaskService:  pack.TaskService,
		Caster:       caster,
	}
	t, err := pack.TaskService.DispatchJob(models.ScanDirectoryTask, meta, nil)
	if err != nil {
		log.ErrTrace(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	t.SetCleanup(
		func(t *task.Task) {
			if caster.IsEnabled() {
				caster.Close()
			}
		},
	)

	w.WriteHeader(http.StatusOK)
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
func getSharedFiles(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}

	shares, err := pack.ShareService.GetFileSharesWithUser(u)
	if SafeErrorAndExit(err, w) {
		return
	}

	var children = make([]*fileTree.WeblensFileImpl, 0)
	for _, share := range shares {
		f, err := pack.FileService.GetFileSafe(share.FileId, u, share)
		if err != nil {
			if errors.Is(err, werror.ErrNoFile) {
				log.Error.Println("Could not find file acompanying a file share")
				continue
			}
			safeErr, code := werror.TrySafeErr(err)
			writeJson(w, code, safeErr)
			return
		}
		children = append(children, f)
	}

	childInfos := make([]rest.FileInfo, 0, len(children))
	for _, child := range children {
		fInfo, err := rest.WeblensFileToFileInfo(child, pack, false)
		if SafeErrorAndExit(err, w) {
			return
		}
		childInfos = append(childInfos, fInfo)
	}

	medias, err := getChildMedias(pack, children)
	if SafeErrorAndExit(err, w) {
		return
	}

	var mediaInfos []rest.MediaInfo
	for _, m := range medias {
		mediaInfos = append(mediaInfos, rest.MediaToMediaInfo(m))
	}

	fakeSelfFile := rest.FileInfo{
		Id:           "shared",
		Filename:     "SHARED",
		IsDir:        true,
		PortablePath: fileTree.NewFilePath("", "SHARED", "").ToPortable() + "/",
	}

	res := rest.FolderInfoResponse{Children: childInfos, Medias: mediaInfos, Self: fakeSelfFile}

	writeJson(w, http.StatusOK, res)
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
func createTakeout(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}
	if u == nil {
		u = pack.UserService.GetPublicUser()
	}

	takeoutRequest, err := readCtxBody[rest.FilesListParams](w, r)
	if err != nil {
		return
	}
	if len(takeoutRequest.FileIds) == 0 {
		writeJson(w, http.StatusBadRequest, rest.WeblensErrorInfo{Error: "Cannot takeout 0 files"})
		return
	}

	share, err := getShareFromCtx[*models.FileShare](w, r)
	if err != nil {
		return
	}

	var files []*fileTree.WeblensFileImpl
	for _, fileId := range takeoutRequest.FileIds {
		file, err := pack.FileService.GetFileSafe(fileId, u, share)
		if SafeErrorAndExit(err, w) {
			return
		}

		files = append(files, file)
	}

	// If we only have 1 file, and it is not a directory, we should have requested to just
	// simply download that file on it's own, not zip it.
	if len(files) == 1 && !files[0].IsDir() {
		writeJson(w, http.StatusBadRequest, rest.WeblensErrorInfo{Error: "Single non-directory file should not be zipped"})
		return
	}

	caster := models.NewSimpleCaster(pack.ClientService)
	meta := models.ZipMeta{
		Files:       files,
		Requester:   u,
		Share:       share,
		Caster:      caster,
		FileService: pack.FileService,
	}
	t, err := pack.TaskService.DispatchJob(models.CreateZipTask, meta, nil)

	if err != nil {
		log.ShowErr(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	completed, status := t.Status()
	if completed && status == task.TaskSuccess {
		res := rest.TakeoutInfo{TakeoutId: t.GetResult("takeoutId").(string), Single: false, Filename: t.GetResult("filename").(string)}
		writeJson(w, http.StatusOK, res)
	} else {
		res := rest.TakeoutInfo{TaskId: t.TaskId()}
		writeJson(w, http.StatusAccepted, res)
	}
}

func getFileStat(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}
	fileId := chi.URLParam(r, "fileId")

	f, err := pack.FileService.GetFileSafe(fileId, u, nil)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	size := f.Size()

	writeJson(w,
		http.StatusOK,
		FileStat{Name: f.Filename(), Size: size, IsDir: f.IsDir(), ModTime: f.ModTime(), Exists: true},
	)
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
func autocompletePath(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	searchPath := r.URL.Query().Get("searchPath")
	if len(searchPath) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}

	lastSlashI := strings.LastIndex(searchPath, "/")
	childName := searchPath[lastSlashI+1:]
	searchPath = searchPath[:lastSlashI] + "/"

	folder, err := pack.FileService.UserPathToFile(searchPath, u)
	if err != nil {
		log.ShowErr(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	children := folder.GetChildren()
	if folder.GetParentId() == "ROOT" {
		trashIndex := slices.IndexFunc(children, func(f *fileTree.WeblensFileImpl) bool {
			return f.ID() == u.TrashId
		})
		children = internal.Banish(children, trashIndex)
	}

	var filenames []string
	for _, child := range children {
		filenames = append(filenames, child.Filename())
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

	var childInfos []rest.FileInfo
	for _, match := range matches {
		f := children[match.OriginalIndex]

		childInfo, err := rest.WeblensFileToFileInfo(f, pack, false)
		if SafeErrorAndExit(err, w) {
			return
		}
		childInfos = append(childInfos, childInfo)
	}

	selfInfo, err := rest.WeblensFileToFileInfo(folder, pack, false)
	if SafeErrorAndExit(err, w) {
		return
	}

	ret := rest.FolderInfoResponse{Children: childInfos, Self: selfInfo}
	writeJson(w, http.StatusOK, ret)
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
//	@Param		request	body		rest.RestoreFilesBody				true	"Restore files request body"
//	@Success	200		{object}	http.restoreFiles.restoreFilesInfo	"Restore files info"
//	@Failure	400
//	@Failure	404
//	@Failure	500
//	@Router		/files/restore [post]
func restoreFiles(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}

	body, err := readCtxBody[rest.RestoreFilesBody](w, r)
	if err != nil {
		return
	}

	if body.Timestamp == 0 {
		writeJson(w, http.StatusBadRequest, rest.WeblensErrorInfo{Error: "Missing body parameter 'timestamp'"})
		return
	}
	restoreTime := time.UnixMilli(body.Timestamp)

	lt := pack.FileService.GetJournalByTree("USERS").Get(body.NewParentId)
	if lt == nil {
		writeJson(w, http.StatusNotFound, rest.WeblensErrorInfo{Error: "Could not find new parent"})
		return
	}

	var newParent *fileTree.WeblensFileImpl
	if lt.GetLatestAction().GetActionType() == fileTree.FileDelete {
		newParent, err = pack.FileService.GetFileSafe(u.HomeId, u, nil)

		// this should never error, but you never know
		if SafeErrorAndExit(err, w) {
			return
		}
	} else {
		newParent, err = pack.FileService.GetFileSafe(body.NewParentId, u, nil)
		if SafeErrorAndExit(err, w) {
			return
		}
	}

	err = pack.FileService.RestoreFiles(body.FileIds, newParent, restoreTime, pack.Caster)
	if SafeErrorAndExit(err, w) {
		return
	}

	type restoreFilesInfo struct {
		NewParentId string `json:"newParentId"`
	} //	@name	RestoreFilesInfo
	res := restoreFilesInfo{NewParentId: newParent.ID()}

	writeJson(w, http.StatusOK, res)
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
//	@Param		request	body	rest.UpdateFileParams	true	"Update file request body"
//	@Success	200
//	@Failure	403
//	@Failure	404
//	@Failure	500
//	@Router		/files/{fileId} [patch]
func updateFile(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}
	fileId := chi.URLParam(r, "fileId")
	updateInfo, err := readCtxBody[rest.UpdateFileParams](w, r)
	if err != nil {
		return
	}

	file, err := pack.FileService.GetFileSafe(fileId, u, nil)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if pack.FileService.IsFileInTrash(file) {
		writeJson(w, http.StatusForbidden, rest.WeblensErrorInfo{Error: "cannot rename file in trash"})
		return
	}

	err = pack.FileService.RenameFile(file, updateInfo.NewName, pack.Caster)
	if SafeErrorAndExit(err, w) {
		return
	}

	w.WriteHeader(http.StatusOK)
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
func moveFiles(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}
	sh, err := getShareFromCtx[*models.FileShare](w, r)
	if err != nil {
		return
	}

	filesData, err := readCtxBody[rest.MoveFilesParams](w, r)
	if err != nil {
		return
	}

	newParent, err := pack.FileService.GetFileSafe(filesData.NewParentId, u, sh)
	if SafeErrorAndExit(err, w) {
		return
	}

	if len(filesData.Files) == 0 {
		writeJson(w, http.StatusBadRequest, rest.WeblensErrorInfo{Error: "No file ids provided"})
		return
	}

	if filesData.NewParentId == "" {
		writeJson(w, http.StatusBadRequest, rest.WeblensErrorInfo{Error: "No parent id provided"})
		return
	}

	var files []*fileTree.WeblensFileImpl
	parentId := ""
	for _, fileId := range filesData.Files {
		f, err := pack.FileService.GetFileSafe(fileId, u, sh)
		if SafeErrorAndExit(err, w) {
			return
		}
		if parentId == "" {
			parentId = f.GetParentId()
		} else if parentId != f.GetParentId() {
			writeJson(w, http.StatusBadRequest, rest.WeblensErrorInfo{Error: "All files must have the same parent"})
			return
		}

		files = append(files, f)
	}

	err = pack.FileService.MoveFiles(files, newParent, "USERS", pack.Caster)
	if SafeErrorAndExit(err, w) {
		return
	}

	w.WriteHeader(http.StatusOK)
}

// TrashFiles godoc
//
//	@ID			TrashFiles
//
//	@Security	SessionAuth
//
//	@Summary	Move a list of files to the trash
//	@Tags		Files
//	@Param		request	body	rest.FilesListParams	true	"Trash files request body"
//	@Success	200
//	@Failure	403
//	@Failure	404
//	@Failure	500
//	@Router		/files/trash [patch]
// func trashFiles(w http.ResponseWriter, r *http.Request) {
// 	pack := getServices(r)
// 	params, err := readCtxBody[rest.FilesListParams](w, r)
// 	if err != nil {
// 		return
// 	}
// 	fileIds := params.FileIds
// 	u, err := getUserFromCtx(r)
// 	if SafeErrorAndExit(err, w) {
// 		return
// 	}
//
// 	var files []*fileTree.WeblensFileImpl
// 	for _, fileId := range fileIds {
// 		file, err := pack.FileService.GetFileSafe(fileId, u, nil)
// 		if SafeErrorAndExit(err, w) {
// 			return
// 		}
//
// 		if u != pack.FileService.GetFileOwner(file) {
// 			writeJson(w, http.StatusForbidden, rest.WeblensErrorInfo{Error: "You can't trash this file"})
// 			return
// 		}
//
// 		if file.GetParentId() == u.TrashId {
// 			writeJson(w, http.StatusBadRequest, rest.WeblensErrorInfo{Error: "File is already in trash"})
// 			return
// 		}
//
// 		files = append(files, file)
// 	}
// 	err = pack.FileService.MoveFilesToTrash(files, u, nil, pack.Caster)
// 	if SafeErrorAndExit(err, w) {
// 		return
// 	}
//
// 	w.WriteHeader(http.StatusOK)
// }

// UnTrashFiles godoc
//
//	@ID			UnTrashFiles
//
//	@Security	SessionAuth
//
//	@Summary	Move a list of files out of the trash, restoring them to where they were before
//	@Tags		Files
//	@Param		request	body	rest.FilesListParams	true	"UnTrash files request body"
//	@Success	200
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/files/untrash [patch]
func unTrashFiles(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}

	params, err := readCtxBody[rest.FilesListParams](w, r)
	if err != nil {
		return
	}
	fileIds := params.FileIds

	caster := models.NewSimpleCaster(pack.ClientService)
	defer caster.Close()

	var files []*fileTree.WeblensFileImpl
	for _, fileId := range fileIds {
		file, err := pack.FileService.GetFileSafe(fileId, u, nil)
		if err != nil || u != pack.FileService.GetFileOwner(file) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		files = append(files, file)
	}

	err = pack.FileService.ReturnFilesFromTrash(files, caster)
	if err != nil {
		log.ShowErr(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteFiles godoc
//
//	@ID			DeleteFiles
//
//	@Security	SessionAuth
//
//	@Summary	Delete Files "permanently"
//	@Tags		Files
//	@Param		request	body	rest.FilesListParams	true	"Delete files request body"
//	@Success	200
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/files [delete]
func deleteFiles(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}
	params, err := readCtxBody[rest.FilesListParams](w, r)
	if err != nil {
		return
	}
	fileIds := params.FileIds

	if len(fileIds) == 0 {
		writeJson(w, http.StatusBadRequest, rest.WeblensErrorInfo{Error: "No file ids provided"})
		return
	}

	var files []*fileTree.WeblensFileImpl
	for _, fileId := range fileIds {
		file, err := pack.FileService.GetFileSafe(fileId, u, nil)
		if SafeErrorAndExit(err, w) {
			return
		} else if u != pack.FileService.GetFileOwner(file) {
			writeJson(w, http.StatusNotFound, rest.WeblensErrorInfo{Error: "Could not find file to delete"})
			return
		} else if !pack.FileService.IsFileInTrash(file) {
			writeJson(w, http.StatusForbidden, rest.WeblensErrorInfo{Error: "Cannot delete file not in trash"})
			return
		}
		files = append(files, file)
	}

	err = pack.FileService.DeleteFiles(files, "USERS", pack.Caster)
	if SafeErrorAndExit(err, w) {
		return
	}

	w.WriteHeader(http.StatusOK)
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
//	@Success	200		{object}	rest.NewUploadInfo		"Upload Info"
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/upload [post]
func newUploadTask(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}

	upInfo, err := readCtxBody[rest.NewUploadParams](w, r)
	if SafeErrorAndExit(err, w) {
		return
	}

	if upInfo.TotalUploadSize == 0 {
		writeJson(w, http.StatusBadRequest, rest.WeblensErrorInfo{Error: "Total upload size cannot be 0"})
		return
	}

	meta := models.UploadFilesMeta{
		ChunkStream:  make(chan models.FileChunk, 10),
		RootFolderId: upInfo.RootFolderId,
		ChunkSize:    upInfo.ChunkSize,
		TotalSize:    upInfo.TotalUploadSize,
		FileService:  pack.FileService,
		MediaService: pack.MediaService,
		TaskService:  pack.TaskService,
		TaskSubber:   pack.ClientService,
		User:         u,
		Caster:       pack.Caster,
	}
	t, err := pack.TaskService.DispatchJob(models.UploadFilesTask, meta, nil)
	if SafeErrorAndExit(err, w) {
		return
	}

	uploadInfo := rest.NewUploadInfo{UploadId: t.TaskId()}
	writeJson(w, http.StatusCreated, uploadInfo)
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
//	@Param		request		body		rest.NewFilesParams	true	"New file params"
//	@Success	200			{object}	rest.NewFilesInfo	"FileIds"
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/upload/{uploadId} [post]
func newFileUpload(w http.ResponseWriter, r *http.Request) {
	uploadTaskId := chi.URLParam(r, "uploadId")
	params, err := readCtxBody[rest.NewFilesParams](w, r)
	if err != nil {
		return
	}

	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}
	uTask := pack.TaskService.GetTask(uploadTaskId)
	if uTask == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	completed, _ := uTask.Status()
	if completed {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var ids []fileTree.FileId
	for _, newFInfo := range params.NewFiles {
		parent, err := pack.FileService.GetFileSafe(newFInfo.ParentFolderId, u, nil)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		child, _ := parent.GetChild(newFInfo.NewFileName)
		if child != nil {
			writeJson(w, http.StatusConflict, rest.WeblensErrorInfo{Error: "File with the same name already exists in folder"})
			return
		}

		err = uTask.Manipulate(
			func(meta task.TaskMetadata) error {
				uploadMeta := meta.(models.UploadFilesMeta)

				newF, err := pack.FileService.CreateFile(parent, newFInfo.NewFileName, uploadMeta.UploadEvent, pack.Caster)
				if err != nil {
					return err
				}

				ids = append(ids, newF.ID())

				uploadMeta.ChunkStream <- models.FileChunk{
					NewFile: newF, ContentRange: "0-0/" + strconv.FormatInt(newFInfo.FileSize, 10),
				}

				return nil
			},
		)

		if SafeErrorAndExit(err, w) {
			return
		}
	}

	newInfo := rest.NewFilesInfo{FileIds: ids}
	writeJson(w, http.StatusCreated, newInfo)
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
//	@Param		chunk		formData	file	true	"File chunk"
//	@Success	200
//	@Failure	401
//	@Failure	404
//	@Failure	500
//	@Router		/upload/{uploadId}/file/{fileId} [put]
func handleUploadChunk(w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	uploadId := chi.URLParam(r, "uploadId")

	t := pack.TaskService.GetTask(uploadId)
	if t == nil {
		writeJson(w, http.StatusNotFound, rest.WeblensErrorInfo{Error: "No upload exists with given id"})
		return
	}

	fileId := chi.URLParam(r, "fileId")

	// We are about to read from the clientConn, which could take a while.
	// Since we actually got this request, we know the clientConn is not abandoning us,
	// so we can safely clear the timeout, which the task will re-enable if needed.
	t.ClearTimeout()

	chunk, err := internal.OracleReader(r.Body, r.ContentLength)
	if err != nil {
		log.ShowErr(err)
		// err = t.AddChunkToStream(fileId, nil, "0-0/-1")
		// if err != nil {
		// 	util.ShowErr(err)
		// }
		writeJson(w, http.StatusInternalServerError, rest.WeblensErrorInfo{Error: "Failed to read request body"})
		return
	}

	ctHeader := r.Header.Get("Content-Type")
	if strings.HasPrefix(ctHeader, "multipart/form-data") {
		boundry := strings.Split(ctHeader, "boundary=")[1]
		chunk = bytes.TrimPrefix(chunk, []byte("--"+boundry+"\r\n"))
		chunk = bytes.TrimSuffix(chunk, []byte("--"+boundry+"--\r\n"))

		chunk = chunk[bytes.Index(chunk, []byte("\r\n\r\n"))+4:]

		// Search for null byte and truncate the chunk
		// var b byte = 0
		// var counter = 0
		// for b != 0 || counter == 0 {
		// 	b = chunk[counter]
		// 	counter++
		// }
		// chunk = chunk[counter:]

	}

	err = t.Manipulate(
		func(meta task.TaskMetadata) error {
			chunkData := models.FileChunk{FileId: fileId, Chunk: chunk, ContentRange: r.Header["Content-Range"][0]}
			meta.(models.UploadFilesMeta).ChunkStream <- chunkData

			return nil
		},
	)

	if err != nil {
		log.ShowErr(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Helper Function
func getChildMedias(pack *models.ServicePack, children []*fileTree.WeblensFileImpl) ([]*models.Media, error) {
	var medias []*models.Media
	for _, child := range children {
		var m *models.Media
		contentId := child.GetContentId()
		if child.IsDir() && contentId == "" {
			coverId, err := pack.FileService.GetFolderCover(child)
			if err != nil {
				return nil, err
			}

			if coverId != "" {
				child.SetContentId(coverId)
				contentId = coverId
			}
		}

		m = pack.MediaService.Get(contentId)

		if m != nil {
			medias = append(medias, m)
		}
	}

	return medias, nil
}

// Helper Function
// Format and write back directory information. Authorization checks should be done before this function
func formatRespondFolderInfo(dir *fileTree.WeblensFileImpl, w http.ResponseWriter, r *http.Request) {
	pack := getServices(r)
	u, err := getUserFromCtx(r)
	if SafeErrorAndExit(err, w) {
		return
	}
	share, err := getShareFromCtx[*models.FileShare](w, r)
	if SafeErrorAndExit(err, w) {
		return
	}

	var parentsInfo []rest.FileInfo
	parent := dir.GetParent()
	for parent.ID() != "ROOT" && pack.AccessService.CanUserAccessFile(u, parent, share) && !pack.FileService.GetFileOwner(parent).IsSystemUser() {
		parentInfo, err := rest.WeblensFileToFileInfo(parent, pack, false)
		if SafeErrorAndExit(err, w) {
			return
		}
		parentsInfo = append(parentsInfo, parentInfo)
		parent = parent.GetParent()
	}

	children := dir.GetChildren()

	mediaFiles := append(children, dir)
	medias, err := getChildMedias(pack, mediaFiles)
	if SafeErrorAndExit(err, w) {
		return
	}

	childInfos := make([]rest.FileInfo, 0, len(children))
	for _, child := range children {
		info, err := rest.WeblensFileToFileInfo(child, pack, false)
		if SafeErrorAndExit(err, w) {
			return
		}
		childInfos = append(childInfos, info)
	}

	selfInfo, err := rest.WeblensFileToFileInfo(dir, pack, false)
	if SafeErrorAndExit(err, w) {
		return
	}

	var mediaInfos []rest.MediaInfo
	for _, m := range medias {
		mediaInfos = append(mediaInfos, rest.MediaToMediaInfo(m))
	}

	packagedInfo := rest.FolderInfoResponse{Self: selfInfo, Children: childInfos, Parents: parentsInfo, Medias: mediaInfos}
	writeJson(w, http.StatusOK, packagedInfo)
}

// Helper Function
func formatRespondPastFolderInfo(folderId fileTree.FileId, pastTime time.Time, w http.ResponseWriter, r *http.Request) {
	log.Trace.Func(func(l log.Logger) {
		l.Printf("Getting past folder [%s] at time [%s]", folderId, pastTime)
	})

	pack := getServices(r)

	pastFile, err := pack.FileService.GetJournalByTree("USERS").GetPastFile(folderId, pastTime)
	if SafeErrorAndExit(err, w) {
		return
	}
	pastFileInfo, err := rest.WeblensFileToFileInfo(pastFile, pack, true)
	if SafeErrorAndExit(err, w) {
		return
	}

	var parentsInfo []rest.FileInfo
	parentId := pastFile.GetParentId()
	if parentId == "" {
		writeJson(w, http.StatusNotFound, rest.WeblensErrorInfo{Error: "Could not find parent folder"})
		return
	}
	for parentId != "ROOT" {
		pastParent, err := pack.FileService.GetJournalByTree("USERS").GetPastFile(parentId, pastTime)
		if SafeErrorAndExit(err, w) {
			return
		}

		parentInfo, err := rest.WeblensFileToFileInfo(pastParent, pack, true)
		if SafeErrorAndExit(err, w) {
			return
		}

		parentsInfo = append(parentsInfo, parentInfo)
		parentId = pastParent.GetParentId()
	}

	children, err := pack.FileService.GetJournalByTree("USERS").GetPastFolderChildren(pastFile, pastTime)
	if SafeErrorAndExit(err, w) {
		return
	}

	childrenInfos := make([]rest.FileInfo, 0, len(children))
	for _, child := range children {
		childInfo, err := rest.WeblensFileToFileInfo(child, pack, true)
		if SafeErrorAndExit(err, w) {
			return
		}
		childrenInfos = append(childrenInfos, childInfo)
	}

	var medias []*models.Media
	for _, child := range children {
		m := pack.MediaService.Get(child.GetContentId())
		if m != nil {
			medias = append(medias, m)
		}
	}

	var mediaInfos []rest.MediaInfo
	for _, m := range medias {
		mediaInfos = append(mediaInfos, rest.MediaToMediaInfo(m))
	}

	packagedInfo := rest.FolderInfoResponse{Self: pastFileInfo, Children: childrenInfos, Parents: parentsInfo, Medias: mediaInfos}
	writeJson(w, http.StatusOK, packagedInfo)
}
