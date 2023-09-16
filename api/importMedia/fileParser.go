package importMedia

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethrousseau/weblens/api/interfaces"
	"github.com/ethrousseau/weblens/api/util"

	"github.com/ethrousseau/weblens/api/database"
)

func HandleNewImage(filepath string, db database.Weblensdb) (*interfaces.Media, error) {
	defer func() {
		if err := recover(); err != nil {
			util.Error.Printf("Recovered panic while parsing new file (%s)", filepath)
		}
	}()

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

	db.DbAddMedia(&m)

	return &m, nil
}

func middleware(path string, d fs.DirEntry, wp util.WorkerPool, db database.Weblensdb) error {
	if d.IsDir() {
		return nil
	}
	if strings.HasSuffix(path, ".DS_Store") || strings.HasSuffix(path, ".thumb.jpeg") {
		return nil
	}

	wp.AddTask(func() {
		HandleNewImage(path, db)
	})

	return nil
}

func ScanDirectory(scanDir string, recursive bool) (util.WorkerPool) {
	scanDir = util.GuaranteeAbsolutePath(scanDir)
	util.Debug.Printf("Beginning directory scan: %s\n", scanDir)
	//start := time.Now()

	db := database.New()

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

	ms := db.GetMediaInDirectory(scanDir)
	for _, m := range ms {
		if !recursive && filepath.Dir(m.Filepath) != scanDir {
			continue
		}
		_, err := os.Stat(m.Filepath)
		if err != nil {
			util.Debug.Println("Remove: ", m.Filepath)
			db.RemoveMediaByFilepath(m.Filepath)
		}
	}

	return wp

	// util.Debug.Printf("Finished file scan in %fs", time.Since(start).Seconds())


}