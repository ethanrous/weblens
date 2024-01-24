package dataProcess

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/barasher/go-exiftool"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
)

// Global exiftool
var gexift *exiftool.Exiftool

func InitGExif(bufSize int64) *exiftool.Exiftool {
	gbuf := make([]byte, int(bufSize))
	et, err := exiftool.NewExiftool(exiftool.Api("largefilesupport"), exiftool.ExtractAllBinaryMetadata(), exiftool.Buffer(gbuf, int(bufSize)))
	if err != nil {
		util.DisplayError(err)
		return nil
	}
	gexift = et

	return gexift
}

func processMediaFile(t *Task) {
	meta := t.metadata.(ScanMetadata)
	m := meta.PartialMedia
	file := meta.File

	if m == nil {
		t.error(errors.New("attempted to process nil media"))
		return
	}

	defer m.Clean()

	var parseAnyway bool = false
	filled, _ := m.IsFilledOut(false)
	if filled && !parseAnyway {
		util.Warning.Println("Tried to process media file that already exists in the database")
		t.success()
		return
	}

	m.FileId = file.Id()

	if m.MediaType == nil {
		err := m.ComputeExif()
		util.FailOnError(err, "Failed to extract exif data")
	}

	m.Id()

	// Files that are not "media" (jpeg, png, mov, etc.) should not be stored in the media database
	if m.MediaType == nil || !m.MediaType.IsDisplayable {
		caster.PushFileUpdate(file)
		t.success()
		return
	}

	thumb, err := m.GenerateThumbnail()
	if err != nil {
		t.error(err)
		return
	}

	if m.BlurHash == "" {
		m.GenerateBlurhash(thumb)
	}

	if m.Owner == "" {
		m.Owner = file.Owner()
	}

	db := dataStore.NewDB()

	err = db.AddMedia(m)
	if err != nil {
		t.error(err)
		return
	}

	m.SetImported()
	file.SetMedia(m)

	caster.PushFileUpdate(file)
	t.success()

}

func initTmpExiftool(bufSize int64) (et *exiftool.Exiftool, err error) {
	buf := make([]byte, int(bufSize))
	et, err = exiftool.NewExiftool(exiftool.Api("largefilesupport"), exiftool.ExtractAllBinaryMetadata(), exiftool.Buffer(buf, int(bufSize)))
	return
}

// We don't pre-load directories that are larger than 5GB
const DIR_PRELOAD_LIMIT = 1000 * 1000 * 1000 * 5

func scanDirectory(t *Task) {
	meta := t.metadata.(ScanMetadata)
	scanDir := meta.File

	if scanDir.Filename() == ".user_trash" {
		caster.PushTaskUpdate(t.taskId, "scan_complete", t.result) // Let any client subscribers know we are done
		t.success()
		return
	}

	recursive := meta.Recursive
	deepScan := meta.DeepScan

	mediaToScan := []*dataStore.WeblensFile{}
	dirsToScan := []*dataStore.WeblensFile{}

	// "deepScan" defines wether or not we should go sync
	// with the real filesystem for this scan. Otherwise,
	// just handle processing media we already know about
	if deepScan {
		scanDir.ReadDir()
	}

	children := scanDir.GetChildren()
	for _, c := range children {
		if !c.IsDir() {
			displayable, err := c.IsDisplayable()
			if displayable && errors.Is(err, dataStore.ErrNoMedia) {
				mediaToScan = append(mediaToScan, c)
			}
		} else if recursive && c != scanDir {
			dirsToScan = append(dirsToScan, c)
		}
	}

	for _, dir := range dirsToScan {
		t := NewTask("scan_directory", ScanMetadata{File: dir, Recursive: recursive, DeepScan: deepScan})
		ttInstance.globalQueue.QueueTask(t)
	}

	if len(mediaToScan) == 0 {
		caster.PushTaskUpdate(t.taskId, "scan_complete", t.result) // Let any client subscribers know we are done
		t.success("No media to scan:", scanDir.String())
		return
	}

	// Wait on all tasks on files as not to be destructive
	util.Each(mediaToScan, func(wf *dataStore.WeblensFile) { util.Each(wf.GetTasks(), func(t dataStore.Task) { t.Wait() }) })

	util.Info.Printf("Beginning directory scan for %s [M %d][R %t][D %t]\n", scanDir.String(), len(mediaToScan), recursive, deepScan)

	sw := util.NewStopwatch(fmt.Sprint("Directory scan ", scanDir.String()))

	var dirSize int64
	var allMetadata []exiftool.FileMetadata
	util.Each(mediaToScan, func(f *dataStore.WeblensFile) { s, err := f.Size(); util.DisplayError(err); dirSize += s })

	var preload bool = false
	if dirSize < DIR_PRELOAD_LIMIT {
		util.Debug.Println("Preloading exifdata for", scanDir.String())
		preload = true
		et, err := initTmpExiftool(dirSize)
		if err != nil {
			t.error(err)

			if et != nil {
				// This might panic...
				et.Close()
			}
			return
		}

		sw.Lap("Initiated exiftool")

		paths := util.Map(mediaToScan, func(f *dataStore.WeblensFile) string { return f.String() })
		allMetadata = et.ExtractMetadata(paths...)

		err = et.Close()
		if err != nil {
			t.error(err)
			return
		}

		sw.Lap("Pre-Loaded Metadata")
	}

	for i, file := range mediaToScan {
		newMedia := &dataStore.Media{}
		newMedia.FileId = file.Id()
		if preload {
			newMedia.DumpRawExif(allMetadata[i].Fields)
		}
		file.SetMedia(newMedia)
	}

	sw.Lap("Loaded medias")

	wq := NewWorkQueue()

	slices.SortFunc(mediaToScan, func(a, b *dataStore.WeblensFile) int { return strings.Compare(a.Filename(), b.Filename()) })

	for _, file := range mediaToScan {
		m, err := file.GetMedia()
		if err != nil {
			util.DisplayError(err)
			continue
		}
		t := NewTask("scan_file", ScanMetadata{File: file, PartialMedia: m})
		wq.QueueTask(t)
	}

	sw.Lap("Queued all tasks")

	wq.SignalAllQueued()
	wq.Wait(true) // Park this thread and wait for the media to finish computing

	sw.Lap("All tasks finished")
	sw.Stop()
	sw.PrintResults()

	caster.PushTaskUpdate(t.taskId, "scan_complete", t.result) // Let any client subscribers know we are done
	t.success()
}
