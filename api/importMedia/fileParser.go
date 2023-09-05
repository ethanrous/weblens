package importMedia

import (
	"io/fs"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	util "github.com/ethrousseau/weblens/api/utils"

	"github.com/ethrousseau/weblens/api/database"
)

func HandleNewImage(filepath string, db database.Weblensdb) () {
	defer func() {
		if err := recover(); err != nil {
			util.Error.Printf("Panic recovered parsing new file (%s): %s\n", filepath, string(debug.Stack()))
		}
	}()

	start := time.Now()

	m, exists := db.GetImageByFilename(filepath)

	var parseAnyway bool = false
	if exists && !parseAnyway {
		return
	}

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
}

func ScanAllMedia() {
	util.Debug.Print("Beginning file scan\n")
	imagesDir := "/Users/ethan/repos/weblens/images"

	db := database.New()

	wp := util.NewWorkerPool(10)
	wp.Run()

	filepath.WalkDir(imagesDir, func(path string, d fs.DirEntry,  err error) error {
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".MP4") || strings.HasSuffix(path, ".DS_Store") {
			return nil
		}


		wp.AddTask(func() {
			HandleNewImage(path, db)
		})

		return nil
	})

	util.Debug.Print("Finished file scan\n")

}