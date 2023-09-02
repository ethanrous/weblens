package importMedia

import (
	"io/fs"
	"path/filepath"
	"sync"

	log "github.com/ethrousseau/weblens/api/utils"

	"github.com/ethrousseau/weblens/api/database"
	"github.com/ethrousseau/weblens/api/interfaces"
)



func HandleNewImage(filepath string, m *interfaces.Media, db database.Weblensdb) () {
	if m.Filepath != "" {
		m.Filepath = filepath
	}

	m.ExtractExif()

	i := m.ReadFile()

	thumb := m.GenerateThumbnail(i)

	if m.BlurHash == "" {
		m.GenerateBlurhash(thumb)
	}

	db.DbAddMedia(m)

}

func ScanAllMedia() {
	log.Debug.Print("Beginning file scan\n")
	imagesDir := "/Users/ethan/repos/weblens/images"

	var wg sync.WaitGroup
	db := database.New()

	filepath.WalkDir(imagesDir, func(path string, d fs.DirEntry,  err error) error {
		if d.IsDir() {
			return nil
		}
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			defer func() {
				if err := recover(); err != nil {
					log.Error.Printf("Panic recovered parsing new file (%s): %s\n", p, err)
				}
			}()

			m, exists := db.GetImageByFilename(p)

			updateAnyway := false

			if !exists || updateAnyway {
				HandleNewImage(p, &m, db)
			}

		}(path)
		return nil
	})

	wg.Wait()

	log.Debug.Print("Finished file scan\n")

}