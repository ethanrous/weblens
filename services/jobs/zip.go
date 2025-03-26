package jobs

import (
	"archive/zip"
	"context"
	"maps"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/models/client"
	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/modules/crypto"
	slices_mod "github.com/ethanrous/weblens/modules/slices"
	"github.com/ethanrous/weblens/task"
	"github.com/pkg/errors"
	"github.com/saracen/fastzip"
)

var ErrEmptyZip = errors.New("zip file is empty")

func CreateZip(t *task.Task) {
	zipMeta := t.GetMeta().(models.ZipMeta)

	if len(zipMeta.Files) == 0 {
		t.ReqNoErr(ErrEmptyZip)
	}

	filesInfoMap := map[string]os.FileInfo{}

	slices_mod.Map(
		zipMeta.Files,
		func(file *file_model.WeblensFileImpl) error {
			return file.RecursiveMap(
				func(f *file_model.WeblensFileImpl) error {
					stat, err := os.Stat(f.AbsPath())
					if err != nil {
						t.ReqNoErr(err)
					}
					filesInfoMap[f.AbsPath()] = stat
					return nil
				},
			)
		},
	)

	paths := slices.Sorted(maps.Keys(filesInfoMap))

	var takeoutKey string
	if len(zipMeta.Files) == 1 {
		takeoutKey = zipMeta.Files[0].Filename()
	} else {
		takeoutKey = crypto.HashString(strings.Join(paths, ""))
	}

	zipName := takeoutKey
	var zipExists bool

	if !strings.HasSuffix(zipName, ".zip") {
		zipName = zipName + ".zip"
	}

	zipFile, err := zipMeta.FileService.NewZip(zipName, zipMeta.Requester)
	if err != nil && strings.Contains(err.Error(), "file already exists") {
		err = nil
		zipExists = true
	} else if err != nil {
		t.ReqNoErr(err)
	}

	if zipExists {
		t.SetResult(task.TaskResult{"takeoutId": zipFile.ID(), "filename": zipFile.Filename()})
		// Let any client subscribers know we are done
		zipMeta.Caster.PushTaskUpdate(t, client.ZipCompleteEvent, t.GetResults())
		t.Success()
		return
	}

	zipMeta.Caster.PushTaskUpdate(t, client.TaskCreatedEvent, task.TaskResult{"totalFiles": len(filesInfoMap)})

	fp, err := os.Create(zipFile.AbsPath())
	if err != nil {
		t.ReqNoErr(err)
	}
	defer func(fp *os.File) {
		err := fp.Close()
		if err != nil {
			t.Log.Error().Stack().Err(err).Msg("")
		}
	}(fp)

	a, err := fastzip.NewArchiver(
		fp, zipMeta.Files[0].GetParent().AbsPath(),
		fastzip.WithStageDirectory(zipFile.GetParent().AbsPath()),
		fastzip.WithArchiverBufferSize(1024*1024*1024),
		fastzip.WithArchiverMethod(zip.Store),
	)

	if err != nil {
		t.ReqNoErr(err)
	}

	defer func(a *fastzip.Archiver) {
		err := a.Close()
		if err != nil {
			t.Log.Error().Stack().Err(err).Msg("")
		}
	}(a)

	var archiveErr *error

	// Shove archive to child thread so we can send updates with main thread
	go func() {
		err := a.Archive(context.Background(), filesInfoMap)
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

			zipMeta.Caster.PushTaskUpdate(
				t, client.ZipProgressEvent, task.TaskResult{
					"completedFiles": int(entries), "totalFiles": totalFiles,
					"bytesSoFar": bytes,
					"bytesTotal": bytesTotal,
					"speedBytes": int((float64(byteDiff) / float64(timeNs)) * float64(time.Second)),
				},
			)
			prevBytes = bytes
			sinceUpdate = 0
		}

		time.Sleep(time.Duration(updateInterval))
	}
	if archiveErr != nil {
		t.ReqNoErr(*archiveErr)
	}

	t.SetResult(task.TaskResult{"takeoutId": zipFile.ID(), "filename": zipFile.Filename()})
	zipMeta.Caster.PushTaskUpdate(
		t, models.ZipCompleteEvent, t.GetResults(),
	) // Let any client subscribers know we are done
	t.Success()
}
