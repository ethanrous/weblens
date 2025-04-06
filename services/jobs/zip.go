package jobs

import (
	"archive/zip"
	"maps"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/ethanrous/weblens/models/client"
	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/job"
	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/modules/crypto"
	slices_mod "github.com/ethanrous/weblens/modules/slices"
	task_mod "github.com/ethanrous/weblens/modules/task"
	"github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/services/context"
	"github.com/pkg/errors"
	"github.com/saracen/fastzip"
)

var ErrEmptyZip = errors.New("zip file is empty")

func CreateZip(tsk task_mod.Task) {
	t := tsk.(*task.Task)

	ctx := t.Ctx.(*context.AppContext)
	zipMeta := t.GetMeta().(job.ZipMeta)

	if len(zipMeta.Files) == 0 {
		t.ReqNoErr(ErrEmptyZip)
	}

	filesInfoMap := map[string]os.FileInfo{}

	slices_mod.Map(
		zipMeta.Files,
		func(file *file_model.WeblensFileImpl) error {
			return file.RecursiveMap(
				func(f *file_model.WeblensFileImpl) error {
					fAbs := f.GetPortablePath().ToAbsolute()
					stat, err := os.Stat(fAbs)
					if err != nil {
						t.ReqNoErr(err)
					}
					filesInfoMap[fAbs] = stat
					return nil
				},
			)
		},
	)

	paths := slices.Sorted(maps.Keys(filesInfoMap))

	var takeoutKey string
	if len(zipMeta.Files) == 1 {
		takeoutKey = zipMeta.Files[0].GetPortablePath().Filename()
	} else {
		takeoutKey = crypto.HashString(strings.Join(paths, ""))
	}

	zipName := takeoutKey
	var zipExists bool

	if !strings.HasSuffix(zipName, ".zip") {
		zipName = zipName + ".zip"
	}

	zipFile, err := ctx.FileService.NewZip(zipName, zipMeta.Requester)
	if err != nil && strings.Contains(err.Error(), "file already exists") {
		err = nil
		zipExists = true
	} else if err != nil {
		t.ReqNoErr(err)
	}

	if zipExists {
		t.SetResult(task_mod.TaskResult{"takeoutId": zipFile.ID(), "filename": zipFile.GetPortablePath().Filename()})
		// Let any client subscribers know we are done
		notif := client.NewTaskNotification(t, websocket.ZipCompleteEvent, t.GetResults())
		ctx.Notify(notif)
		t.Success()
		return
	}

	notif := client.NewTaskNotification(t, websocket.TaskCreatedEvent, task_mod.TaskResult{"totalFiles": len(filesInfoMap)})
	ctx.Notify(notif)
	t.Success()

	fp, err := os.Create(zipFile.GetPortablePath().ToAbsolute())
	if err != nil {
		t.ReqNoErr(err)
	}
	defer func(fp *os.File) {
		err := fp.Close()
		if err != nil {
			t.Ctx.Log().Error().Stack().Err(err).Msg("")
		}
	}(fp)

	a, err := fastzip.NewArchiver(
		fp, zipMeta.Files[0].GetParent().GetPortablePath().ToAbsolute(),
		fastzip.WithStageDirectory(zipFile.GetParent().GetPortablePath().ToAbsolute()),
		fastzip.WithArchiverBufferSize(1024*1024*1024),
		fastzip.WithArchiverMethod(zip.Store),
	)

	if err != nil {
		t.ReqNoErr(err)
	}

	defer func(a *fastzip.Archiver) {
		err := a.Close()
		if err != nil {
			t.Ctx.Log().Error().Stack().Err(err).Msg("")
		}
	}(a)

	var archiveErr *error

	// Shove archive to child thread so we can send updates with main thread
	go func() {
		err := a.Archive(t.Ctx, filesInfoMap)
		if err != nil {
			archiveErr = &err
		}
	}()

	bytesTotal := slices_mod.Reduce(
		zipMeta.Files, func(file *file_model.WeblensFileImpl, acc int64) int64 {
			num := file.Size()
			return acc + num
		}, 0,
	)

	var entries int64
	var bytes int64
	var prevBytes int64 = -1
	var sinceUpdate int64 = 0
	totalFiles := len(filesInfoMap)

	const updateInterval = 500 * int64(time.Millisecond)

	// Update client over websocket until entire archive has been written, or an error is thrown
	for int64(totalFiles) > entries {
		if archiveErr != nil {
			break
		}
		sinceUpdate++
		bytes, entries = a.Written()
		if bytes != prevBytes {
			byteDiff := bytes - prevBytes
			timeNs := updateInterval * sinceUpdate

			notif := client.NewTaskNotification(t, websocket.ZipProgressEvent,
				task_mod.TaskResult{
					"completedFiles": int(entries), "totalFiles": totalFiles,
					"bytesSoFar": bytes,
					"bytesTotal": bytesTotal,
					"speedBytes": int((float64(byteDiff) / float64(timeNs)) * float64(time.Second)),
				})
			t.Ctx.Notify(notif)
			prevBytes = bytes
			sinceUpdate = 0
		}

		time.Sleep(time.Duration(updateInterval))
	}
	if archiveErr != nil {
		t.ReqNoErr(*archiveErr)
	}

	t.SetResult(task_mod.TaskResult{"takeoutId": zipFile.ID(), "filename": zipFile.GetPortablePath().Filename()})

	notif = client.NewTaskNotification(t, websocket.ZipCompleteEvent, t.GetResults())
	ctx.Notify(notif)

	t.Success()
}
