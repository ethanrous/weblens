package importMedia

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/util"
)

func HandleNewImage(filepath string, db dataStore.Weblensdb) (*dataStore.Media, error) {
	defer func() {
		if err := recover(); err != nil {
			util.Error.Printf("Recovered panic while parsing new file (%s): %s", filepath, err)
		}
	}()

	// util.Debug.Println("Starting parse of file: ", filepath)

	m := db.GetMediaByFilepath(filepath, true)

	var parseAnyway bool = false
	if m.IsFilledOut(false) && !parseAnyway {
		return &m, nil
	}

	if m.Filepath == "" {
		m.Filepath = filepath
	}

	if m.MediaType.FriendlyName == "" {
		err := m.ExtractExif()
		if err != nil {
			return nil, err
		}
	}

	if m.MediaType.FriendlyName != "File" {
		thumb := m.GenerateThumbnail()
		if m.BlurHash == "" {
			m.GenerateBlurhash(thumb)
		}
	}

	m.GenerateFileHash()

	// util.Debug.Println("Finished parse of file: ", filepath)

	db.DbAddMedia(&m)

	return &m, nil
}

func middleware(path string, d fs.DirEntry, wp util.WorkerPool, db dataStore.Weblensdb) error {
	if d.IsDir() {
		return nil
	}
	if strings.HasSuffix(path, ".DS_Store") || strings.HasSuffix(path, ".thumb.jpeg") {
		return nil
	}

	wp.AddTask(func() {
		_, err:= HandleNewImage(path, db)
		util.DisplayError(err, "Error attempting to import new media")
	})

	return nil
}

func ScanDirectory(scanDir string, recursive bool) (util.WorkerPool) {
	scanDir = dataStore.GuaranteeAbsolutePath(scanDir)
	util.Debug.Printf("Beginning directory scan: %s\n", scanDir)
	//start := time.Now()

	db := dataStore.NewDB()

	wp := util.NewWorkerPool(10)
	wp.Run()

	if recursive {
		filepath.WalkDir(scanDir, func(path string, d fs.DirEntry,  err error) error {
			errr := middleware(path, d, wp, db)
			return errr
		})
	} else {
		files, err := os.ReadDir(scanDir)
		if err != nil {
			panic(err)
		}
		for _, d := range files {
			middleware(filepath.Join(scanDir, d.Name()), d, wp, db)
		}
	}

	ms := db.GetMediaInDirectory(scanDir, recursive)

	for _, m := range ms {
		_, err := os.Stat(dataStore.GuaranteeAbsolutePath(m.Filepath))
		if errors.Is(err, os.ErrNotExist) {
			util.Error.Println("ERR: ", err)
			util.Debug.Println("Remove: ", m.Filepath)
			db.RemoveMediaByFilepath(m.Filepath)
		}
	}

	return wp

	// util.Debug.Printf("Finished file scan in %fs", time.Since(start).Seconds())


}