package dataProcess

import (
	"errors"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
)

func ProcessMediaFile(file *dataStore.WeblensFileDescriptor, m *dataStore.Media, db *dataStore.Weblensdb) (err error) {
	defer util.RecoverPanic("Panic caught while processing media file %s", file.String())

	if m == nil {
		return errors.New("attempted to process nil media")
	}

	defer m.Clean()

	var parseAnyway bool = false
	filled, _ := m.IsFilledOut(false)
	if filled && !parseAnyway {
		// util.Warning.Println("Tried to process media file that already exists in the database")
		return
	}

	m.FileId = file.Id()

	if m.MediaType.FriendlyName == "" {
		err := m.ComputeExif()
		util.FailOnError(err, "Failed to extract exif data")
	}

	m.FileHash = util.HashOfString(8, file.String())
	// Files that are not "media" (jpeg, png, mov, etc.) should not be stored in the media database
	if !m.MediaType.IsDisplayable {
		// PushItemCreate(file)
		return
	}

	thumb, err := m.GenerateThumbnail()
	if err != nil {
		return
	}

	if m.BlurHash == "" {
		m.GenerateBlurhash(thumb)
	}

	if m.Owner == "" {
		m.Owner = file.Owner()
	}

	db.DbAddMedia(m)
	file.SetMedia(m)

	// PushItemCreate(file)

	return nil
}

func scanDirectory(t *task) {
	meta := t.metadata.(ScanMetadata)
	scanDir := meta.File

	if scanDir.Filename() == ".user_trash" {
		t.BroadcastComplete("scan_complete")
		return
	}

	recursive := meta.Recursive
	deepScan := meta.DeepScan

	util.Info.Printf("Beginning directory scan (recursive: %t, deep: %t): %s\n", recursive, deepScan, scanDir.String())

	mediaToScan := []*dataStore.WeblensFileDescriptor{}
	dirsToScan := []*dataStore.WeblensFileDescriptor{}

	if deepScan {
		scanDir.ReadDir()
	}

	children := scanDir.GetChildren()
	for _, c := range children {
		if !c.IsDir() {
			if m, _ := c.GetMedia(); m == nil && c.IsDisplayable() {
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
		t.BroadcastComplete("scan_complete")
		util.Info.Printf("No media to scan: %s", scanDir.String())
		return
	}

	et, err := exiftool.NewExiftool(exiftool.Api("largefilesupport"))
	if err != nil {
		util.DisplayError(err)
		t.err = err
	}

	paths := util.Map(mediaToScan, func(f *dataStore.WeblensFileDescriptor) string { return f.String() })
	allMetadata := et.ExtractMetadata(paths...)
	for i, file := range mediaToScan {
		newMedia := &dataStore.Media{}
		newMedia.FileId = file.Id()
		newMedia.DumpRawExif(allMetadata[i].Fields)
		file.SetMedia(newMedia)
	}

	wq := NewWorkQueue()

	start := time.Now()
	jpgFromRaws := util.Filter(mediaToScan, func(f *dataStore.WeblensFileDescriptor) bool {
		if m, err := f.GetMedia(); err != nil {
			util.DisplayError(err)
			return false
		} else if m.QueryExif("JpgFromRawLength") != nil {
			return true
		}
		return false
	})
	if len(jpgFromRaws) != 0 {
		jpgFromRawsTask := NewTask("preload_meta", PreloadMetaMeta{Files: jpgFromRaws, ExifThumbType: "JpgFromRaw"})
		wq.QueueTask(jpgFromRawsTask)
	}

	previewImages := util.Filter(mediaToScan, func(f *dataStore.WeblensFileDescriptor) bool {
		if m, err := f.GetMedia(); err != nil {
			util.DisplayError(err)
			return false
		} else if m.QueryExif("PreviewImageLength") != nil {
			return true
		}
		return false
	})
	if len(previewImages) != 0 {
		previewImagesTask := NewTask("preload_meta", PreloadMetaMeta{Files: previewImages, ExifThumbType: "PreviewImage"})
		wq.QueueTask(previewImagesTask)
	}

	// We must signal to the work queue that all tasks have
	// been queued before waiting, otherwise we will never wake up
	wq.SignalAllQueued()
	wq.Wait()

	util.Debug.Println("Pre-import meta collection took", time.Since(start))

	// Tell the wq we want to add more tasks
	wq.ClearAllQueued()

	start = time.Now()
	for _, file := range mediaToScan {
		m, err := file.GetMedia()
		if err != nil {
			util.DisplayError(err)
			continue
		}
		t := NewTask("scan_file", ScanMetadata{File: file, PartialMedia: m})
		wq.QueueTask(t)
	}

	wq.SignalAllQueued()
	wq.Wait() // Park this thread and wait for the media to finish computing
	util.Debug.Printf("Completed scanning %d images in %v", len(mediaToScan), time.Since(start))

	t.BroadcastComplete("scan_complete") // Let any client subscribers know we are done
}
