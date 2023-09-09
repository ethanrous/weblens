package importMedia

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/ethrousseau/weblens/api/interfaces"
	util "github.com/ethrousseau/weblens/api/utils"

	"github.com/ethrousseau/weblens/api/database"
)

func HandleNewImage(filepath string, db database.Weblensdb) (*interfaces.Media) {
	defer func() {
		if err := recover(); err != nil {
			util.Error.Printf("Panic recovered parsing new file (%s):\n%s\n\n%s\n", filepath, err, string(debug.Stack()))
		}
	}()

	start := time.Now()

	m, exists := db.GetMediaByFilepath(filepath)

	var parseAnyway bool = false
	if exists && !parseAnyway {
		return &m
	}

	util.Debug.Printf("Starting file scan for %s", filepath)

	if m.Filepath == "" {
		m.Filepath = filepath
	}

	if m.MediaType.FriendlyName == "" {
		m.ExtractExif()
	}

	i := m.ReadFile()

	thumb := m.GenerateThumbnail(i)

	if m.BlurHash == "" {

		m.GenerateBlurhash(thumb)
	}

	db.DbAddMedia(&m)

	util.Debug.Printf("File synced with database in %dms (%s)", time.Since(start).Milliseconds(), filepath)

	return &m
}

func middleware(path string, d fs.DirEntry, wp util.WorkerPool, db database.Weblensdb) error {
	if d.IsDir() {
		return nil
	}
	if strings.HasSuffix(path, ".DS_Store") {
		return nil
	}

	wp.AddTask(func() {
		HandleNewImage(path, db)
	})

	return nil
}

func ScanAllMedia(scanDir string, recursive bool) {
	util.Debug.Print("Beginning file scan\n")

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

	util.Debug.Print("Finished file scan\n")

}