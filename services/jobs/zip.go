package jobs

import (
	"archive/zip"
	"maps"
	"os"
	"slices"
	"strings"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/job"
	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/modules/crypto"
	"github.com/ethanrous/weblens/modules/errors"
	slices_mod "github.com/ethanrous/weblens/modules/slices"
	task_mod "github.com/ethanrous/weblens/modules/task"
	"github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/services/context"
	file_service "github.com/ethanrous/weblens/services/file"
	"github.com/ethanrous/weblens/services/notify"
	"github.com/saracen/fastzip"
)

// ErrEmptyZip error reporting that a zip file is empty
var ErrEmptyZip = errors.New("zip file is empty")

// CreateZip creates a zip archive from the files specified in the task metadata.
func CreateZip(tsk task_mod.Task) {
	t := tsk.(*task.Task)

	ctx, ok := context.FromContext(t.Ctx)
	if !ok {
		t.Fail(errors.New("context is not a RequestContext"))

		return
	}

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
		takeoutKey = crypto.HashString(strings.Join(paths, ""))[:8]
	}

	zipName := takeoutKey

	if !strings.HasSuffix(zipName, ".zip") {
		zipName = zipName + ".zip"
	}

	ctx = ctx.WithValue(file_service.SkipJournalKey, true)

	zipFile, err := ctx.FileService.GetZip(ctx, zipName)
	if err == nil {
		t.SetResult(task_mod.Result{"takeoutID": zipFile.ID(), "filename": zipFile.GetPortablePath().Filename()})
		// Let any client subscribers know we are done
		notif := notify.NewTaskNotification(t, websocket.ZipCompleteEvent, t.GetResults())
		ctx.Notify(ctx, notif)
		t.Success()

		return
	}

	zipFile, err = ctx.FileService.NewZip(ctx, zipName, zipMeta.Requester)
	if err != nil {
		t.Fail(err)

		return
	}

	notif := notify.NewTaskNotification(t, websocket.TaskCreatedEvent, task_mod.Result{"totalFiles": len(filesInfoMap)})
	ctx.Notify(ctx, notif)

	zw, err := zipFile.Writer()
	if err != nil {
		t.Fail(err)

		return
	}

	defer zw.Close() //nolint:errcheck

	a, err := fastzip.NewArchiver(
		zw, zipMeta.Files[0].GetParent().GetPortablePath().ToAbsolute(),
		fastzip.WithStageDirectory(zipFile.GetParent().GetPortablePath().ToAbsolute()),
		fastzip.WithArchiverBufferSize(1024*1024*1024),
		fastzip.WithArchiverMethod(zip.Store),
	)
	if err != nil {
		t.Fail(err)

		return
	}

	defer a.Close() //nolint:errcheck

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

	var (
		entries     int64
		bytes       int64
		prevBytes   int64 = -1
		sinceUpdate int64
	)

	totalFiles := len(filesInfoMap)

	const updateInterval = 500 * int64(time.Millisecond)

	t.OnResult(func(result task_mod.Result) {
		notif := notify.NewTaskNotification(t, websocket.ZipProgressEvent, result)
		ctx.Notify(ctx, notif)
	})

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

			t.SetResult(task_mod.Result{
				"completedFiles": int(entries), "totalFiles": totalFiles,
				"bytesSoFar": bytes,
				"bytesTotal": bytesTotal,
				"speedBytes": int((float64(byteDiff) / float64(timeNs)) * float64(time.Second)),
			})

			prevBytes = bytes
			sinceUpdate = 0
		}

		time.Sleep(time.Duration(updateInterval))
	}

	if archiveErr != nil {
		t.Fail(*archiveErr)

		return
	}

	t.ClearOnResult()

	t.SetResult(task_mod.Result{
		"takeoutID":      zipFile.ID(),
		"filename":       zipFile.GetPortablePath().Filename(),
		"completedFiles": totalFiles,
		"bytesSoFar":     bytesTotal,
	})

	notif = notify.NewTaskNotification(t, websocket.ZipCompleteEvent, t.GetResults())
	ctx.Notify(ctx, notif)

	t.Success()
}
