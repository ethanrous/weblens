package job

import (
	"encoding/json"
	"fmt"
	"hash"
	"slices"
	"sync"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	media_model "github.com/ethanrous/weblens/models/media"
	share_model "github.com/ethanrous/weblens/models/share"
	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/models/tower"
	user_model "github.com/ethanrous/weblens/models/usermodel"

	"github.com/ethanrous/weblens/modules/wlerrors"
	slices_mod "github.com/ethanrous/weblens/modules/wlslices"
	"github.com/rs/zerolog/log"
)

// TaskSubscriber provides methods for managing task subscriptions.
type TaskSubscriber interface {
	FolderSubToTask(folderID string, taskID string)
	UnsubTask(taskID string)
	// TaskSubToPool(taskID string, poolID string)
}

// TaskDispatcher provides methods for dispatching jobs to task pools.
type TaskDispatcher interface {
	DispatchJob(jobName string, meta task.Metadata, pool *task.Pool) (*task.Task, error)
}

// ScanMeta holds metadata for file scanning tasks.
type ScanMeta struct {
	File         *file_model.WeblensFileImpl
	PartialMedia *media_model.Media

	FileBytes []byte
}

// MetaString returns a JSON string representation of the scan metadata.
func (m ScanMeta) MetaString() string {
	data := map[string]any{
		"JobName": ScanFileTask,
		"FileIds": m.File.ID(),
	}

	bs, err := json.Marshal(data)
	if err != nil {
		err = wlerrors.WithStack(err)
		log.Error().Stack().Err(err).Msg("Could not marshal scan metadata")

		return ""
	}

	return string(bs)
}

// FormatToResult converts the scan metadata to a task result.
func (m ScanMeta) FormatToResult() task.Result {
	return task.Result{
		"filename": m.File.GetPortablePath(),
	}
}

// JobName returns the appropriate job name based on whether the file is a directory.
func (m ScanMeta) JobName() string {
	if m.File.IsDir() {
		return ScanDirectoryTask
	}

	return ScanFileTask
}

// Verify checks that the scan metadata contains all required fields.
func (m ScanMeta) Verify() error {
	if m.File == nil {
		return wlerrors.New("no file in scan metadata")
	}

	return nil
}

// ZipMeta holds metadata for zip file creation tasks.
type ZipMeta struct {
	Share     *share_model.FileShare
	Requester *user_model.User

	Files []*file_model.WeblensFileImpl
}

// MetaString returns a string representation of the zip metadata.
func (m ZipMeta) MetaString() string {
	ids := slices_mod.Map(
		m.Files, func(f *file_model.WeblensFileImpl) string {
			return f.ID()
		},
	)

	slices.Sort(ids)

	idsString, err := json.Marshal(ids)
	if err != nil {
		err = wlerrors.WithStack(err)
		log.Error().Stack().Err(err).Msg("could not marshal zip metadata")

		return ""
	}

	var shareBit string
	if m.Share != nil {
		shareBit = string(m.Share.ShareID.Hex()) + m.Share.LastUpdated().String()
	}

	data := CreateZipTask + string(idsString) + string(m.Requester.GetUsername()) + shareBit

	return data
}

// FormatToResult converts the zip metadata to a task result.
func (m ZipMeta) FormatToResult() task.Result {
	return task.Result{
		"filenames": slices_mod.Map(
			m.Files, func(f *file_model.WeblensFileImpl) string {
				return f.GetPortablePath().ToPortable()
			},
		),
	}
}

// JobName returns the job name for zip creation tasks.
func (m ZipMeta) JobName() string {
	return CreateZipTask
}

// Verify checks that the zip metadata contains all required fields.
func (m ZipMeta) Verify() error {
	if len(m.Files) == 0 {
		return wlerrors.New("no files in zip metadata")
	} else if m.Requester == nil {
		return wlerrors.New("no requester in zip metadata")
	}

	if m.Share == nil {
		log.Warn().Msg("No share in zip meta...")
	}

	return nil
}

// MoveMeta holds metadata for file move tasks.
type MoveMeta struct {
	User                *user_model.User
	DestinationFolderID string
	NewFilename         string
	FileIDs             []string
}

// MetaString returns a JSON string representation of the move metadata.
func (m MoveMeta) MetaString() string {
	data := map[string]any{
		"JobName":     MoveFileTask,
		"FileIds":     m.FileIDs,
		"DestID":      m.DestinationFolderID,
		"NewFileName": m.NewFilename,
	}

	bs, err := json.Marshal(data)
	if err != nil {
		err = wlerrors.WithStack(err)
		log.Error().Stack().Err(err).Msg("could not marshal move metadata")

		return ""
	}

	return string(bs)
}

// FormatToResult converts the move metadata to a task result.
func (m MoveMeta) FormatToResult() task.Result {
	return task.Result{}
}

// JobName returns the job name for file move tasks.
func (m MoveMeta) JobName() string {
	return MoveFileTask
}

// Verify checks that the move metadata contains all required fields.
func (m MoveMeta) Verify() error {
	return nil
}

// FileChunk represents a chunk of file data during upload.
type FileChunk struct {
	NewFile      *file_model.WeblensFileImpl
	FileID       string
	ContentRange string

	Chunk []byte
}

// UploadFilesMeta holds metadata for file upload tasks.
type UploadFilesMeta struct {
	// TaskService *task.WorkerPool
	// TaskSubber  TaskSubscriber
	ChunkStream chan FileChunk

	User         *user_model.User
	Share        *share_model.FileShare
	RootFolderID string
	ChunkSize    int64
}

// MetaString returns a string representation of the upload metadata.
func (m UploadFilesMeta) MetaString() string {
	return fmt.Sprintf("%s%s%d", UploadFilesTask, m.RootFolderID, m.ChunkSize)
}

// FormatToResult converts the upload metadata to a task result.
func (m UploadFilesMeta) FormatToResult() task.Result {
	return task.Result{}
}

// JobName returns the job name for file upload tasks.
func (m UploadFilesMeta) JobName() string {
	return UploadFilesTask
}

// Verify checks that the upload metadata contains all required fields.
func (m UploadFilesMeta) Verify() error {
	if m.ChunkStream == nil {
		return wlerrors.New("no chunk stream in upload metadata")
	} else if m.RootFolderID == "" {
		return wlerrors.New("no root folder in upload metadata")
	} else if m.ChunkSize == 0 {
		return wlerrors.New("no chunk size in upload metadata")
	} else if m.User == nil {
		return wlerrors.New("no user in upload metadata")
	}

	return nil
}

// FsStatMeta holds metadata for filesystem statistics gathering tasks.
type FsStatMeta struct {
	RootDir *file_model.WeblensFileImpl
}

// MetaString returns a JSON string representation of the filesystem statistics metadata.
func (m FsStatMeta) MetaString() string {
	data := map[string]any{
		"JobName":    GatherFsStatsTask,
		"RootFolder": m.RootDir.ID(),
	}

	bs, err := json.Marshal(data)
	if err != nil {
		err = wlerrors.WithStack(err)
		log.Error().Stack().Err(err).Msg("could not marshal fs stat metadata")

		return ""
	}

	return string(bs)
}

// FormatToResult converts the filesystem statistics metadata to a task result.
func (m FsStatMeta) FormatToResult() task.Result {
	return task.Result{}
}

// JobName returns the job name for filesystem statistics gathering tasks.
func (m FsStatMeta) JobName() string {
	return GatherFsStatsTask
}

// Verify checks that the filesystem statistics metadata contains all required fields.
func (m FsStatMeta) Verify() error {
	return nil
}

// FileUploadProgress tracks the progress of a file upload operation.
type FileUploadProgress struct {
	Hash          hash.Hash
	File          *file_model.WeblensFileImpl
	BytesWritten  int64
	FileSizeTotal int64
}

// BackupMeta holds metadata for backup tasks.
type BackupMeta struct {
	Core tower.Instance
}

// MetaString returns a JSON string representation of the backup metadata.
func (m BackupMeta) MetaString() string {
	data := map[string]any{
		"JobName":  BackupTask,
		"remoteID": m.Core,
	}

	bs, err := json.Marshal(data)
	if err != nil {
		err = wlerrors.WithStack(err)
		log.Error().Stack().Err(err).Msg("could not marshal backup metadata")

		return ""
	}

	return string(bs)
}

// FormatToResult converts the backup metadata to a task result.
func (m BackupMeta) FormatToResult() task.Result {
	return task.Result{}
}

// JobName returns the job name for backup tasks.
func (m BackupMeta) JobName() string {
	return BackupTask
}

// Verify checks that the backup metadata contains all required fields.
func (m BackupMeta) Verify() error {
	if m.Core.TowerID == "" {
		return wlerrors.New("no core id in backup metadata")
	}

	return nil
}

// HashFileMeta holds metadata for file hashing tasks.
type HashFileMeta struct {
	File *file_model.WeblensFileImpl
}

// MetaString returns a JSON string representation of the hash metadata.
func (m HashFileMeta) MetaString() string {
	data := map[string]any{
		"JobName": HashFileTask,
		"fileID":  m.File.ID(),
	}

	bs, err := json.Marshal(data)
	if err != nil {
		err = wlerrors.WithStack(err)
		log.Error().Stack().Err(err).Msg("Could not marshal hasher metadata")
	}

	return string(bs)
}

// FormatToResult converts the hash metadata to a task result.
func (m HashFileMeta) FormatToResult() task.Result {
	return task.Result{}
}

// JobName returns the job name for file hashing tasks.
func (m HashFileMeta) JobName() string {
	return HashFileTask
}

// Verify checks that the hash metadata contains all required fields.
func (m HashFileMeta) Verify() error {
	return nil
}

// LoadFilesystemMeta holds metadata for filesystem loading tasks.
type LoadFilesystemMeta struct {
	File *file_model.WeblensFileImpl
}

// MetaString returns a JSON string representation of the filesystem loading metadata.
func (m LoadFilesystemMeta) MetaString() string {
	data := map[string]any{
		"JobName": LoadFilesystemTask,
		"fileID":  m.File.ID(),
	}

	bs, err := json.Marshal(data)
	if err != nil {
		err = wlerrors.WithStack(err)
		log.Error().Stack().Err(err).Msg("Could not marshal hasher metadata")
	}

	return string(bs)
}

// FormatToResult converts the filesystem loading metadata to a task result.
func (m LoadFilesystemMeta) FormatToResult() task.Result {
	return task.Result{}
}

// JobName returns the job name for filesystem loading tasks.
func (m LoadFilesystemMeta) JobName() string {
	return LoadFilesystemTask
}

// Verify checks that the filesystem loading metadata contains all required fields.
func (m LoadFilesystemMeta) Verify() error {
	return nil
}

// BackupCoreFileMeta holds metadata for backup core file copying tasks.
type BackupCoreFileMeta struct {
	CoreFileID string
	File       *file_model.WeblensFileImpl
	Core       tower.Instance
	Filename   string
}

// MetaString returns a JSON string representation of the backup core file metadata.
func (m BackupCoreFileMeta) MetaString() string {
	data := map[string]any{
		"JobName":    CopyFileFromCoreTask,
		"backupFile": m.File.ID(),
	}

	bs, err := json.Marshal(data)
	if err != nil {
		err = wlerrors.WithStack(err)
		log.Error().Stack().Err(err).Msg("could not marshal backup core metadata")

		return ""
	}

	return string(bs)
}

// FormatToResult converts the backup core file metadata to a task result.
func (m BackupCoreFileMeta) FormatToResult() task.Result {
	return task.Result{}
}

// JobName returns the job name for backup core file copying tasks.
func (m BackupCoreFileMeta) JobName() string {
	return CopyFileFromCoreTask
}

// Verify checks that the backup core file metadata contains all required fields.
func (m BackupCoreFileMeta) Verify() error {
	if m.Core.TowerID == "" {
		return wlerrors.New("no core id in backup core metadata")
	}

	if m.File == nil {
		return wlerrors.New("no file in backup core metadata")
	}

	return nil
}

// RestoreCoreMeta holds metadata for core restoration tasks.
type RestoreCoreMeta struct {
	Core  *tower.Instance
	Local *tower.Instance
}

// MetaString returns a JSON string representation of the restore core metadata.
func (m RestoreCoreMeta) MetaString() string {
	data := map[string]any{
		"JobName": RestoreCoreTask,
	}

	bs, err := json.Marshal(data)
	if err != nil {
		err = wlerrors.WithStack(err)
		log.Error().Stack().Err(err).Msg("could not marshal restore core metadata")

		return ""
	}

	return string(bs)
}

// FormatToResult converts the restore core metadata to a task result.
func (m RestoreCoreMeta) FormatToResult() task.Result {
	return task.Result{}
}

// JobName returns the job name for core restoration tasks.
func (m RestoreCoreMeta) JobName() string {
	return RestoreCoreTask
}

// Verify checks that the restore core metadata contains all required fields.
func (m RestoreCoreMeta) Verify() error {
	if m.Core == nil {
		return wlerrors.New("no core in restore core metadata")
	}

	if m.Local == nil {
		return wlerrors.New("no local in restore core metadata")
	}

	return nil
}

// TaskStage represents a single stage in a multi-stage task.
type TaskStage struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Started  int64  `json:"started"`
	Finished int64  `json:"finished"`

	index int
}

// TaskStages manages the stages of a multi-stage task.
type TaskStages struct {
	data       map[string]TaskStage
	inProgress string
	mu         sync.Mutex
}

// NewBackupTaskStages creates a new TaskStages instance initialized with backup task stages.
func NewBackupTaskStages() *TaskStages {
	return &TaskStages{
		data: map[string]TaskStage{
			"connecting":           {Key: "connecting", Name: "Connecting to Remote", index: 0},
			"fetching_backup_data": {Key: "fetching_backup_data", Name: "Fetching Backup Data", index: 1},
			"writing_users":        {Key: "writing_users", Name: "Writing Users", index: 2},
			"writing_keys":         {Key: "writing_keys", Name: "Writing API Keys", index: 4},
			"writing_instances":    {Key: "writing_instances", Name: "Writing Instances", index: 6},
			"sync_journal":         {Key: "sync_journal", Name: "Calculating New File History", index: 7},
			"sync_fs":              {Key: "sync_fs", Name: "Sync Filesystem", index: 8},
		},
	}
}

// StartStage marks a task stage as started and finishes any currently in-progress stage.
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

// FinishStage marks a task stage as finished.
func (ts *TaskStages) FinishStage(key string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	stage := ts.data[key]

	stage.Finished = time.Now().UnixMilli()
	ts.data[key] = stage
}

// MarshalJSON converts the task stages to JSON format.
func (ts *TaskStages) MarshalJSON() ([]byte, error) {
	ts.mu.Lock()

	data := make([]TaskStage, 0, len(ts.data))
	for _, stage := range ts.data {
		data = append(data, stage)
	}

	ts.mu.Unlock()

	slices.SortFunc(data, func(i, j TaskStage) int { return i.index - j.index })

	return json.Marshal(data)
}
