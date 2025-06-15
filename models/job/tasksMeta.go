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
	user_model "github.com/ethanrous/weblens/models/user"

	"github.com/ethanrous/weblens/modules/errors"
	slices_mod "github.com/ethanrous/weblens/modules/slices"
	task_mod "github.com/ethanrous/weblens/modules/task"
	"github.com/rs/zerolog/log"
)

type TaskSubscriber interface {
	FolderSubToTask(folderId string, taskId string)
	UnsubTask(taskId string)
	// TaskSubToPool(taskId string, poolId string)
}

type TaskDispatcher interface {
	DispatchJob(jobName string, meta task_mod.TaskMetadata, pool *task.TaskPool) (*task.Task, error)
}

type ScanMeta struct {
	File         *file_model.WeblensFileImpl
	PartialMedia *media_model.Media

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

func (m ScanMeta) FormatToResult() task_mod.TaskResult {
	return task_mod.TaskResult{
		"filename": m.File.GetPortablePath(),
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
		return errors.New("no file in scan metadata")
	}

	return nil
}

type ZipMeta struct {
	Share     *share_model.FileShare
	Requester *user_model.User

	Files []*file_model.WeblensFileImpl
}

func (m ZipMeta) MetaString() string {
	ids := slices_mod.Map(
		m.Files, func(f *file_model.WeblensFileImpl) string {
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
		shareBit = string(m.Share.ShareId.Hex()) + m.Share.LastUpdated().String()
	}

	data := CreateZipTask + string(idsString) + string(m.Requester.GetUsername()) + shareBit
	return data
}

func (m ZipMeta) FormatToResult() task_mod.TaskResult {
	return task_mod.TaskResult{
		"filenames": slices_mod.Map(
			m.Files, func(f *file_model.WeblensFileImpl) string {
				return f.GetPortablePath().ToPortable()
			},
		),
	}
}

func (m ZipMeta) JobName() string {
	return CreateZipTask
}

func (m ZipMeta) Verify() error {
	if len(m.Files) == 0 {
		return errors.New("no files in zip metadata")
	} else if m.Requester == nil {
		return errors.New("no requester in zip metadata")
	}

	if m.Share == nil {
		log.Warn().Msg("No share in zip meta...")
	}

	return nil
}

type MoveMeta struct {
	User                *user_model.User
	DestinationFolderId string
	NewFilename         string
	FileIds             []string
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

func (m MoveMeta) FormatToResult() task_mod.TaskResult {
	return task_mod.TaskResult{}
}

func (m MoveMeta) JobName() string {
	return MoveFileTask
}

func (m MoveMeta) Verify() error {
	return nil
}

type FileChunk struct {
	NewFile      *file_model.WeblensFileImpl
	FileId       string
	ContentRange string

	Chunk []byte
}

type UploadFilesMeta struct {
	// TaskService *task.WorkerPool
	// TaskSubber  TaskSubscriber
	ChunkStream chan FileChunk

	User         *user_model.User
	Share        *share_model.FileShare
	RootFolderId string
	ChunkSize    int64
}

func (m UploadFilesMeta) MetaString() string {
	return fmt.Sprintf("%s%s%d", UploadFilesTask, m.RootFolderId, m.ChunkSize)
}

func (m UploadFilesMeta) FormatToResult() task_mod.TaskResult {
	return task_mod.TaskResult{}
}

func (m UploadFilesMeta) JobName() string {
	return UploadFilesTask
}

func (m UploadFilesMeta) Verify() error {
	if m.ChunkStream == nil {
		return errors.New("no chunk stream in upload metadata")
	} else if m.RootFolderId == "" {
		return errors.New("no root folder in upload metadata")
	} else if m.ChunkSize == 0 {
		return errors.New("no chunk size in upload metadata")
	} else if m.User == nil {
		return errors.New("no user in upload metadata")
	}

	return nil
}

type FsStatMeta struct {
	RootDir *file_model.WeblensFileImpl
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

func (m FsStatMeta) FormatToResult() task_mod.TaskResult {
	return task_mod.TaskResult{}
}

func (m FsStatMeta) JobName() string {
	return GatherFsStatsTask
}

func (m FsStatMeta) Verify() error {
	return nil
}

type FileUploadProgress struct {
	Hash          hash.Hash
	File          *file_model.WeblensFileImpl
	BytesWritten  int64
	FileSizeTotal int64
}

type BackupMeta struct {
	Core tower.Instance
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

func (m BackupMeta) FormatToResult() task_mod.TaskResult {
	return task_mod.TaskResult{}
}

func (m BackupMeta) JobName() string {
	return BackupTask
}

func (m BackupMeta) Verify() error {
	if m.Core.TowerId == "" {
		return errors.New("no core id in backup metadata")
	}

	return nil
}

type HashFileMeta struct {
	File *file_model.WeblensFileImpl
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

func (m HashFileMeta) FormatToResult() task_mod.TaskResult {
	return task_mod.TaskResult{}
}

func (m HashFileMeta) JobName() string {
	return HashFileTask
}

func (m HashFileMeta) Verify() error {
	return nil
}

type LoadFilesystemMeta struct {
	File *file_model.WeblensFileImpl
}

func (m LoadFilesystemMeta) MetaString() string {
	data := map[string]any{
		"JobName": LoadFilesystemTask,
		"fileId":  m.File.ID(),
	}

	bs, err := json.Marshal(data)
	if err != nil {
		err = errors.WithStack(err)
		log.Error().Stack().Err(err).Msg("Could not marshal hasher metadata")
	}

	return string(bs)
}

func (m LoadFilesystemMeta) FormatToResult() task_mod.TaskResult {
	return task_mod.TaskResult{}
}

func (m LoadFilesystemMeta) JobName() string {
	return LoadFilesystemTask
}

func (m LoadFilesystemMeta) Verify() error {
	return nil
}

type BackupCoreFileMeta struct {
	CoreFileId string
	File       *file_model.WeblensFileImpl
	Core       tower.Instance
	Filename   string
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

func (m BackupCoreFileMeta) FormatToResult() task_mod.TaskResult {
	return task_mod.TaskResult{}
}

func (m BackupCoreFileMeta) JobName() string {
	return CopyFileFromCoreTask
}

func (m BackupCoreFileMeta) Verify() error {
	if m.Core.TowerId == "" {
		return errors.New("no core id in backup core metadata")
	}

	if m.File == nil {
		return errors.New("no file in backup core metadata")
	}

	return nil
}

type RestoreCoreMeta struct {
	Core  *tower.Instance
	Local *tower.Instance
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

func (m RestoreCoreMeta) FormatToResult() task_mod.TaskResult {
	return task_mod.TaskResult{}
}

func (m RestoreCoreMeta) JobName() string {
	return RestoreCoreTask
}

func (m RestoreCoreMeta) Verify() error {
	if m.Core == nil {
		return errors.New("no core in restore core metadata")
	}
	if m.Local == nil {
		return errors.New("no local in restore core metadata")
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
