package http

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal/werror"
	"github.com/ethanrous/weblens/models"
	"github.com/rs/zerolog/log"
)

type InMemoryFS struct {
	routes   map[string]*memFileReal
	index    *memFileReal
	routesMu *sync.RWMutex
	Pack     *models.ServicePack

	proxyAddress string
}

func (fs *InMemoryFS) loadIndex(uiDir string) string {
	indexPath := filepath.Join(uiDir, "index.html")
	fs.index = readFile(indexPath, fs)
	if !fs.index.exists {
		panic(werror.Errorf("Could not find index file at %s", indexPath))
	}

	return indexPath
}

// Open Implements FileSystem interface
func (fs *InMemoryFS) Open(name string) (http.File, error) {
	log.Trace().Msgf("MemFs Opening file: %s", name)

	if name == "/index" {
		return fs.Index("/index"), nil
	}
	var f *memFileReal
	var ok bool
	name = filepath.Join("./ui/dist/", name)
	fs.routesMu.RLock()
	if f, ok = fs.routes[name]; ok && f.exists {
		fs.routesMu.RUnlock()
		return newWrapFile(f), nil
	} else if !ok {
		fs.routesMu.RUnlock()
		f = readFile(name, fs)
		if f != nil {
			fs.routesMu.Lock()
			fs.routes[f.Name] = f
			fs.routesMu.Unlock()
		}
	} else {
		fs.routesMu.RUnlock()
	}

	// var err error
	if f == nil || !f.exists {
		return fs.Index(name), nil
	}
	return newWrapFile(f), nil
}

func newWrapFile(real *memFileReal) *MemFileWrap {
	return &MemFileWrap{
		at:       0,
		realFile: real,
	}
}

func (fs *InMemoryFS) Exists(prefix string, path string) bool {
	if path == "/" || path == "//" {
		return false
	} else if path == "/index" {
		return true
	}
	_, err := fs.Open(path)
	return err == nil
}

type MemFileWrap struct {
	realFile *memFileReal
	at       int64
}

type memFileReal struct {
	fs     *InMemoryFS
	Name   string
	data   []byte
	exists bool
}

func (mf *memFileReal) Copy() *memFileReal {
	return &memFileReal{
		Name:   mf.Name,
		data:   mf.data,
		exists: mf.exists,
		fs:     mf.fs,
	}
}

type indexFields struct {
	Title       string
	Description string
	Url         string
	Image       string
	Type        string
	VideoUrl    string
	SecureUrl   string
	VideoType   string
}

func (fs *InMemoryFS) Index(loc string) *MemFileWrap {
	index := newWrapFile(fs.index.Copy())
	locIndex := strings.Index(loc, "ui/dist/")
	if locIndex != -1 {
		loc = loc[locIndex+len("ui/dist/"):]
	}

	fields := getIndexFields(loc, fs.proxyAddress, fs.Pack)

	tmpl, err := template.New("index").Parse(string(index.realFile.data))
	if err != nil {
		log.Error().Stack().Err(err).Msg("")
		return index
	}

	buf := bytes.NewBuffer(nil)
	err = tmpl.Execute(buf, fields)
	if err != nil {
		log.Error().Stack().Err(err).Msg("")
		return index
	}

	index.realFile.data = buf.Bytes()

	return index
}

func getIndexFields(path, proxyAddress string, pack *models.ServicePack) indexFields {
	var fields indexFields
	var hasImage bool

	if path[0] == '/' {
		path = path[1:]
	}
	fields.Url = fmt.Sprintf("%s/%s", proxyAddress, path)

	if strings.HasPrefix(path, "files/share/") {
		path = path[len("files/share/"):]
		slashIndex := strings.Index(path, "/")
		if slashIndex == -1 {
			return fields
		}

		shareId := models.ShareId(path[:slashIndex])
		share := pack.ShareService.Get(shareId)
		if share != nil {
			if !share.IsPublic() {
				fields.Title = "Sign in to view"
				fields.Description = "Private file share"
				fields.Image = "/logo_1200.png"
				return fields
			}

			f, err := pack.FileService.GetFileSafe(
				fileTree.FileId(share.GetItemId()), pack.UserService.GetRootUser(), nil,
			)
			if err != nil {
				log.Error().Stack().Err(err).Msg("")
				return fields
			}
			m := pack.MediaService.Get(models.ContentId(f.GetContentId()))
			if f != nil {
				if f.IsDir() {
					cover, err := pack.FileService.GetFolderCover(f)
					if err != nil {
						log.Error().Stack().Err(err).Msg("")
						return fields
					}
					if cover != "" {
						m = pack.MediaService.Get(cover)
					} else {
						imgUrl := fmt.Sprintf("%s/api/static/folder.png", proxyAddress)
						hasImage = true
						fields.Image = imgUrl
					}
				}

				fields.Title = f.Filename()
				fields.Description = "Weblens file share"

				if m != nil {
					if !pack.MediaService.GetMediaType(m).Video {
						imgUrl := fmt.Sprintf(
							"%s/api/media/%s.webp?quality=thumbnail&shareId=%s", proxyAddress,
							f.GetContentId(), share.ID(),
						)
						hasImage = true
						fields.Image = imgUrl
					}
				}
			}
		}
	} else if strings.HasPrefix(path, "albums/") {
		albumId := models.AlbumId(path[len("albums/"):])
		album := pack.AlbumService.Get(albumId)
		if album != nil {
			media := pack.MediaService.Get(album.GetCover())
			if media != nil {
				imgUrl := fmt.Sprintf("%s/api/media/%s.png", proxyAddress, media.ID())
				hasImage = true
				fields.Image = imgUrl
			}
			fields.Title = album.GetName()
			fields.Description = "Weblens album share"
		}
	}

	if !hasImage {
		fields.Image = "/logo_1200.png"
	}

	return fields
}

func readFile(filePath string, fs *InMemoryFS) *memFileReal {
	fdata, err := os.ReadFile(filePath)

	return &memFileReal{
		Name:   filePath,
		data:   fdata,
		exists: err == nil,
		fs:     fs,
	}
}

// Close Implements the comm.File interface
func (f *MemFileWrap) Close() error {
	return nil
}
func (f *MemFileWrap) Stat() (os.FileInfo, error) {
	return &InMemoryFileInfo{f.realFile}, nil
}
func (f *memFileReal) Stat() (os.FileInfo, error) {
	return &InMemoryFileInfo{f}, nil
}
func (f *MemFileWrap) Readdir(count int) ([]os.FileInfo, error) {
	res := make([]os.FileInfo, len(f.realFile.fs.routes))
	i := 0
	for _, file := range f.realFile.fs.routes {
		res[i], _ = file.Stat()
		i++
	}
	return res, nil
}
func (f *MemFileWrap) Read(b []byte) (int, error) {
	n := copy(b, f.realFile.data[f.at:f.at+int64(len(b))])
	f.at += int64(len(b))
	return n, nil
	// buf := bytes.NewBuffer(b)
	// return buf.Write(f.data)

	// i := 0
	// for f.at < int64(len(f.data)) && i < len(b) {
	// 	b[i] = f.data[f.at]
	// 	i++
	// 	f.at++
	// }
	// return i, nil
}
func (f *MemFileWrap) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0:
		f.at = offset
	case 1:
		f.at += offset
	case 2:
		f.at = int64(len(f.realFile.data)) + offset
	}
	return f.at, nil
}

type InMemoryFileInfo struct {
	file *memFileReal
}

// Name Implements os.FileInfo
func (s *InMemoryFileInfo) Name() string       { return s.file.Name }
func (s *InMemoryFileInfo) Size() int64        { return int64(len(s.file.data)) }
func (s *InMemoryFileInfo) Mode() os.FileMode  { return os.ModeTemporary }
func (s *InMemoryFileInfo) ModTime() time.Time { return time.Time{} }
func (s *InMemoryFileInfo) IsDir() bool        { return false }
func (s *InMemoryFileInfo) Sys() interface{}   { return nil }
