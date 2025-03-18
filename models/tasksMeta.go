package models

import (
	"encoding/json"
	"fmt"
	"hash"
	"slices"
	"sync"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/task"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

const (
	ScanDirectoryTask    = "scan_directory"
	ScanFileTask         = "scan_file"
	MoveFileTask         = "move_file"
	UploadFilesTask      = "write_file"
	CreateZipTask        = "create_zip"
	GatherFsStatsTask    = "gather_filesystem_stats"
	BackupTask           = "do_backup"
	HashFileTask         = "hash_file"
	CopyFileFromCoreTask = "copy_file_from_core"
	RestoreCoreTask      = "restore_core"
)

type TaskSubscriber interface {
	FolderSubToTask(folderId fileTree.FileId, taskId task.Id)
	UnsubTask(taskId task.Id)
	// TaskSubToPool(taskId task.Id, poolId task.Id)
}

type TaskDispatcher interface {
	DispatchJob(jobName string, meta task.TaskMetadata, pool *task.TaskPool) (*task.Task, error)
}

type ScanMeta struct {
	FileService  FileService
	MediaService MediaService
	TaskService  TaskDispatcher
	TaskSubber   TaskSubscriber
	Caster       FileCaster
	File         *fileTree.WeblensFileImpl
	PartialMedia *Media

	FileBytes []byte
}

// MetaString overrides marshal function for scanMetadata
func (m ScanMeta) MetaString() string {
	data := map[string]any{
		"JobName": ScanFileTask,
		"FileIds": m.File.ID(),
	}
	bs, err := json.Marshal(data)
	if err != nil {
		err = errors.WithStack(err)
		log.Error().Stack().Err(err).Msg("Could not marshal scan metadata")
		return ""
	}

	return string(bs)
}

func (m ScanMeta) FormatToResult() task.TaskResult {
	return task.TaskResult{
		"filename": m.File.Filename(),
	}
}

func (m ScanMeta) JobName() string {
	if m.File.IsDir() {
		return ScanDirectoryTask
	} else {
		return ScanFileTask
	}
}

func (m ScanMeta) Verify() error {
	if m.File == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "File")
	}
	if m.FileService == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "FileService")
	}
	if m.MediaService == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "MediaService")
	}

	return nil
}

type ZipMeta struct {
	FileService FileService
	Caster      FileCaster
	Share       *FileShare
	Requester   *User

	Files []*fileTree.WeblensFileImpl
}

func (m ZipMeta) MetaString() string {
	ids := internal.Map(
		m.Files, func(f *fileTree.WeblensFileImpl) fileTree.FileId {
			return f.ID()
		},
	)

	slices.Sort(ids)
	idsString, err := json.Marshal(ids)
	if err != nil {
		err = errors.WithStack(err)
		log.Error().Stack().Err(err).Msg("could not marshal zip metadata")
		return ""
	}

	var shareBit string
	if m.Share != nil {
		shareBit = string(m.Share.ShareId) + m.Share.LastUpdated().String()
	}

	data := CreateZipTask + string(idsString) + string(m.Requester.GetUsername()) + shareBit
	return data
}

func (m ZipMeta) FormatToResult() task.TaskResult {
	return task.TaskResult{
		"filenames": internal.Map(
			m.Files, func(f *fileTree.WeblensFileImpl) string {
				return f.Filename()
			},
		),
	}
}

func (m ZipMeta) JobName() string {
	return CreateZipTask
}

func (m ZipMeta) Verify() error {
	if len(m.Files) == 0 {
		return werror.ErrBadJobMetadata(m.JobName(), "files")
	} else if m.Requester == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "requester")
	} else if m.FileService == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "fileService")
	} else if m.Caster == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "caster")
	}

	if m.Share == nil {
		log.Warn().Msg("No share in zip meta...")
	}

	return nil
}

type MoveMeta struct {
	Caster FileCaster

	FileService FileService
	FileEvent   *fileTree.FileEvent

	User                *User
	DestinationFolderId fileTree.FileId
	NewFilename         string
	FileIds             []fileTree.FileId
}

func (m MoveMeta) MetaString() string {
	data := map[string]any{
		"JobName":     MoveFileTask,
		"FileIds":     m.FileIds,
		"DestId":      m.DestinationFolderId,
		"NewFileName": m.NewFilename,
	}
	bs, err := json.Marshal(data)
	if err != nil {
		err = errors.WithStack(err)
		log.Error().Stack().Err(err).Msg("could not marshal move metadata")
		return ""
	}

	return string(bs)
}

func (m MoveMeta) FormatToResult() task.TaskResult {
	return task.TaskResult{}
}

func (m MoveMeta) JobName() string {
	return MoveFileTask
}

func (m MoveMeta) Verify() error {
	return nil
}

type FileChunk struct {
	NewFile      *fileTree.WeblensFileImpl
	FileId       fileTree.FileId
	ContentRange string

	Chunk []byte
}

type UploadFilesMeta struct {
	FileService  FileService
	MediaService MediaService
	TaskService  task.TaskService
	TaskSubber   TaskSubscriber
	Caster       FileCaster
	ChunkStream  chan FileChunk

	UploadEvent *fileTree.FileEvent

	User         *User
	Share        *FileShare
	RootFolderId fileTree.FileId
	ChunkSize    int64
}

func (m UploadFilesMeta) MetaString() string {
	return fmt.Sprintf("%s%s%d", UploadFilesTask, m.RootFolderId, m.ChunkSize)
}

func (m UploadFilesMeta) FormatToResult() task.TaskResult {
	return task.TaskResult{}
}

func (m UploadFilesMeta) JobName() string {
	return UploadFilesTask
}

func (m UploadFilesMeta) Verify() error {
	if m.ChunkStream == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "chunkStream")
	} else if m.RootFolderId == "" {
		return werror.ErrBadJobMetadata(m.JobName(), "rootFolderId")
	} else if m.ChunkSize == 0 {
		return werror.ErrBadJobMetadata(m.JobName(), "chunkSize")
	} else if m.FileService == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "fileService")
	} else if m.MediaService == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "mediaService")
	} else if m.TaskService == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "taskService")
	} else if m.TaskSubber == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "taskSubscriber")
	} else if m.User == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "user")
	} else if m.Caster == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "caster")
	}

	return nil
}

type FsStatMeta struct {
	RootDir *fileTree.WeblensFileImpl
}

func (m FsStatMeta) MetaString() string {
	data := map[string]any{
		"JobName":    GatherFsStatsTask,
		"RootFolder": m.RootDir.ID(),
	}
	bs, err := json.Marshal(data)
	if err != nil {
		err = errors.WithStack(err)
		log.Error().Stack().Err(err).Msg("could not marshal fs stat metadata")
		return ""
	}

	return string(bs)
}

func (m FsStatMeta) FormatToResult() task.TaskResult {
	return task.TaskResult{}
}

func (m FsStatMeta) JobName() string {
	return GatherFsStatsTask
}

func (m FsStatMeta) Verify() error {
	return nil
}

type FileUploadProgress struct {
	Hash          hash.Hash
	File          *fileTree.WeblensFileImpl
	BytesWritten  int64
	FileSizeTotal int64
}

type BackupMeta struct {
	Core             *Instance
	FileService      FileService
	UserService      UserService
	WebsocketService ClientManager
	InstanceService  InstanceService
	TaskService      task.TaskService
	AccessService    AccessService
	Caster           Broadcaster
}

func (m BackupMeta) MetaString() string {
	data := map[string]any{
		"JobName":  BackupTask,
		"remoteId": m.Core,
	}
	bs, err := json.Marshal(data)
	if err != nil {
		err = errors.WithStack(err)
		log.Error().Stack().Err(err).Msg("could not marshal backup metadata")
		return ""
	}

	return string(bs)
}

func (m BackupMeta) FormatToResult() task.TaskResult {
	return task.TaskResult{}
}

func (m BackupMeta) JobName() string {
	return BackupTask
}

func (m BackupMeta) Verify() error {
	if m.Core == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "RemoteId")
	} else if m.FileService == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "FileService")
	} else if m.UserService == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "UserService")
	} else if m.WebsocketService == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "WebsocketService")
	} else if m.InstanceService == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "InstanceService")
	} else if m.TaskService == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "TaskService")
	} else if m.Caster == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "Caster")
	} else if m.AccessService == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "AccessService")
	}

	return nil
}

type HashFileMeta struct {
	File   *fileTree.WeblensFileImpl
	Caster Broadcaster
}

func (m HashFileMeta) MetaString() string {
	data := map[string]any{
		"JobName": HashFileTask,
		"fileId":  m.File.ID(),
	}
	bs, err := json.Marshal(data)
	if err != nil {
		err = errors.WithStack(err)
		log.Error().Stack().Err(err).Msg("Could not marshal hasher metadata")
	}

	return string(bs)
}

func (m HashFileMeta) FormatToResult() task.TaskResult {
	return task.TaskResult{}
}

func (m HashFileMeta) JobName() string {
	return HashFileTask
}

func (m HashFileMeta) Verify() error {
	return nil
}

type BackupCoreFileMeta struct {
	FileService FileService
	CoreFileId  string
	File        *fileTree.WeblensFileImpl
	Caster      Broadcaster
	Core        *Instance
	Filename    string
}

func (m BackupCoreFileMeta) MetaString() string {
	data := map[string]any{
		"JobName":    CopyFileFromCoreTask,
		"backupFile": m.File.ID(),
	}

	bs, err := json.Marshal(data)
	if err != nil {
		err = errors.WithStack(err)
		log.Error().Stack().Err(err).Msg("could not marshal backup core metadata")
		return ""
	}

	return string(bs)
}

func (m BackupCoreFileMeta) FormatToResult() task.TaskResult {
	return task.TaskResult{}
}

func (m BackupCoreFileMeta) JobName() string {
	return CopyFileFromCoreTask
}

func (m BackupCoreFileMeta) Verify() error {
	if m.Core == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "Core")
	}
	if m.File == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "File")
	}
	if m.FileService == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "FileService")
	}

	return nil
}

type RestoreCoreMeta struct {
	Core  *Instance
	Local *Instance
	Pack  *ServicePack
}

func (m RestoreCoreMeta) MetaString() string {
	data := map[string]any{
		"JobName": RestoreCoreTask,
	}

	bs, err := json.Marshal(data)
	if err != nil {
		err = errors.WithStack(err)
		log.Error().Stack().Err(err).Msg("could not marshal restore core metadata")
		return ""
	}

	return string(bs)
}

func (m RestoreCoreMeta) FormatToResult() task.TaskResult {
	return task.TaskResult{}
}

func (m RestoreCoreMeta) JobName() string {
	return RestoreCoreTask
}

func (m RestoreCoreMeta) Verify() error {
	if m.Core == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "Core")
	}
	if m.Local == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "Local")
	}
	if m.Pack == nil {
		return werror.ErrBadJobMetadata(m.JobName(), "ServicePack")
	}

	return nil
}

type TaskStage struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Started  int64  `json:"started"`
	Finished int64  `json:"finished"`

	index int
}

type TaskStages struct {
	data       map[string]TaskStage
	inProgress string
	mu         sync.Mutex
}

func (ts *TaskStages) StartStage(key string) {
	if ts.inProgress != "" {
		ts.FinishStage(ts.inProgress)
	}

	ts.mu.Lock()
	defer ts.mu.Unlock()

	stage := ts.data[key]
	stage.Started = time.Now().UnixMilli()
	ts.data[key] = stage

	ts.inProgress = key
}

func (ts *TaskStages) FinishStage(key string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	stage := ts.data[key]

	stage.Finished = time.Now().UnixMilli()
	ts.data[key] = stage
}

func (ts *TaskStages) MarshalJSON() ([]byte, error) {
	ts.mu.Lock()

	var data []TaskStage
	for _, stage := range ts.data {
		data = append(data, stage)
	}

	ts.mu.Unlock()

	slices.SortFunc(data, func(i, j TaskStage) int { return i.index - j.index })
	return json.Marshal(data)
}

func NewBackupTaskStages() *TaskStages {
	return &TaskStages{
		data: map[string]TaskStage{
			"connecting":           {Key: "connecting", Name: "Connecting to Remote", index: 0},
			"fetching_backup_data": {Key: "fetching_backup_data", Name: "Fetching Backup Data", index: 1},
			"writing_users":        {Key: "writing_users", Name: "Writing Users", index: 2},
			"writing_keys":         {Key: "writing_keys", Name: "Writing Api Keys", index: 4},
			"writing_instances":    {Key: "writing_instances", Name: "Writing Instances", index: 6},
			"sync_journal":         {Key: "sync_journal", Name: "Calculating New File History", index: 7},
			"sync_fs":              {Key: "sync_fs", Name: "Sync Filesystem", index: 8},
		},
	}
}
