package http

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal/env"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
)

type InMemoryFS struct {
	routes   map[string]*memFileReal
	index    *memFileReal
	routesMu *sync.RWMutex
	Pack     *models.ServicePack
}

func (fs *InMemoryFS) loadIndex() string {
	indexPath := filepath.Join(env.GetUIPath(), "index.html")
	fs.index = readFile(indexPath, fs)
	if !fs.index.exists {
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		abs, _ := filepath.Abs(".")
		log.Error.Println("PWD", filepath.Dir(ex), abs)
		panic(werror.Errorf("Could not find index file at %s", indexPath))
	}

	return indexPath
}

// Open Implements FileSystem interface
func (fs *InMemoryFS) Open(name string) (http.File, error) {
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
	at       int64
	realFile *memFileReal
}

type memFileReal struct {
	Name   string
	data   []byte
	exists bool
	fs     *InMemoryFS
}

func (mf *memFileReal) Copy() *memFileReal {
	return &memFileReal{
		Name:   mf.Name,
		data:   mf.data,
		exists: mf.exists,
		fs:     mf.fs,
	}
}

func addIndexTag(tagName, toAdd, content string) string {
	subStr := fmt.Sprintf("og:%s\" content=\"", tagName)
	index := strings.Index(content, subStr)
	if index == -1 {
		log.Error.Println("Failed to find tag", tagName)
		return content
	}
	index += len(subStr)
	return content[:index] + toAdd + content[index:]
}

func (fs *InMemoryFS) Index(loc string) *MemFileWrap {
	index := newWrapFile(fs.index.Copy())
	locIndex := strings.Index(loc, "ui/dist/")
	if locIndex != -1 {
		loc = loc[locIndex+len("ui/dist/"):]
	}

	data := addIndexTag("url", fmt.Sprintf("%s/%s", env.GetHostURL(), loc), string(index.realFile.data))

	fields := getIndexFields(loc, fs.Pack)
	for _, field := range fields {
		data = addIndexTag(field.tag, field.content, data)
	}

	index.realFile.data = []byte(data)

	return index
}

type indexField struct {
	tag     string
	content string
}

func getIndexFields(path string, pack *models.ServicePack) []indexField {

	var fields []indexField
	var hasImage bool

	if strings.HasPrefix(path, "files/share/") {
		path = path[len("files/share/"):]
		slashIndex := strings.Index(path, "/")
		if slashIndex == -1 {
			log.Debug.Println("Could not find slash in path:", path)
			return fields
		}

		shareId := models.ShareId(path[:slashIndex])
		share := pack.ShareService.Get(shareId)
		if share != nil {
			f, err := pack.FileService.GetFileSafe(
				fileTree.FileId(share.GetItemId()), pack.UserService.GetRootUser(), nil,
			)
			if err != nil {
				log.ErrTrace(err)
				return fields
			}
			m := pack.MediaService.Get(models.ContentId(f.GetContentId()))
			if f != nil {
				if f.IsDir() {
					imgUrl := fmt.Sprintf("%s/api/static/folder.png", env.GetHostURL())
					hasImage = true
					fields = append(
						fields, indexField{
							tag:     "image",
							content: imgUrl,
						},
					)
				}

				fields = append(
					fields, indexField{
						tag:     "title",
						content: f.Filename(),
					},
				)

				fields = append(
					fields, indexField{
						tag:     "description",
						content: "Weblens file share",
					},
				)
				if m != nil {
					if !pack.MediaService.GetMediaType(m).IsVideo() {
						imgUrl := fmt.Sprintf(
							"%s/api/media/%s/thumbnail.png?shareId=%s", env.GetHostURL(),
							f.GetContentId(), share.ID(),
						)
						hasImage = true
						fields = append(
							fields, indexField{
								tag:     "image",
								content: imgUrl,
							},
						)
						// videoUrl := fmt.Sprintf("%s/api/media/%s/stream", env.GetHostURL(), f.GetContentId())
						// fields = append(
						// 	fields, indexField{
						// 		tag:     "type",
						// 		content: "video.other",
						// 	},
						// )
						// fields = append(
						// 	fields, indexField{
						// 		tag:     "video",
						// 		content: videoUrl,
						// 	},
						// )
						fields = append(
							fields, indexField{
								tag:     "description",
								content: "Weblens file share",
							},
						)
						// fields = append(
						// 	fields, indexField{
						// 		tag:     "video:type",
						// 		content: "text/html",
						// 	},
						// )
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
				imgUrl := fmt.Sprintf("%s/api/media/%s/thumbnail.png", env.GetHostURL(), media.ID())
				hasImage = true
				fields = append(
					fields, indexField{
						tag:     "image",
						content: imgUrl,
					},
				)
			}

			fields = append(
				fields, indexField{
					tag:     "title",
					content: album.GetName(),
				},
			)

			fields = append(
				fields, indexField{
					tag:     "description",
					content: "Weblens album share",
				},
			)
		}
	}

	if !hasImage {
		// imgUrl := fmt.Sprintf("%s/logo.png", util.GetHostURL())
		fields = append(
			fields, indexField{
				tag:     "image",
				content: "/logo_1200.png",
			},
		)
	}
	// fields = append(
	// 	fields, indexField{
	// 		tag:     "canonical",
	// 		content: util.GetHostURL(),
	// 	},
	// )

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
