package routes

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/ethrousseau/weblens/api/util/wlog"
)

type InMemoryFS struct {
	routes   map[string]*memFileReal
	index    *memFileReal
	routesMu *sync.RWMutex
}

func (fs *InMemoryFS) loadIndex() string {
	indexPath, err := filepath.Abs("../../ui/dist/index.html")
	if err != nil {
		panic("Could not find index file")
	}
	fs.index = readFile(indexPath, fs)
	if !fs.index.exists {
		panic("Could not find index file")
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
	name = filepath.Join("../../ui/dist/", name)
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
		wlog.Error.Println("Failed to find tag", tagName)
		return content
	}
	index += len(subStr)
	return content[:index] + toAdd + content[index:]
}

func (fs *InMemoryFS) Index(loc string) *MemFileWrap {
	index := newWrapFile(fs.index.Copy())
	locIndex := strings.Index(loc, "../../ui/dist/")
	if locIndex != -1 {
		loc = loc[locIndex+len("../../ui/dist/"):]
	}

	data := addIndexTag("url", fmt.Sprintf("%s/%s", util.GetHostURL(), loc), string(index.realFile.data))

	if strings.HasPrefix(loc, "files/share/") {
		loc = loc[len("files/share/"):]
		slashIndex := strings.Index(loc, "/")
		if slashIndex == -1 {
			wlog.Debug.Println("Could not find slash in path:", loc)
			return index
		}
		
		shareId := types.ShareId(loc[:slashIndex])
		share := types.SERV.ShareService.Get(shareId)
		if share != nil {
			f := types.SERV.FileTree.Get(types.FileId(share.GetItemId()))
			m := types.SERV.MediaRepo.Get(f.GetContentId())
			if f != nil {
				if f.IsDir() {
					imgUrl := fmt.Sprintf("%s/api/static/folder.png", util.GetHostURL())
					data = addIndexTag("image", imgUrl, data)
				}

				data = addIndexTag("title", f.Filename(), data)
				data = addIndexTag("description", "Weblens file share", data)
				if m != nil {
					if m.GetMediaType().IsVideo() {
						imgUrl := fmt.Sprintf("%s/api/media/%s/thumbnail.webp", util.GetHostURL(), f.GetContentId())
						data = addIndexTag("image", imgUrl, data)
						videoUrl := fmt.Sprintf("%s/api/media/%s/stream", util.GetHostURL(), f.GetContentId())
						data = addIndexTag("type", "video.other", data)
						data = addIndexTag("video", videoUrl, data)
						data = addIndexTag("video:secure_url", videoUrl, data)
						data = addIndexTag("video:type", "text/html", data)
					}
				}
			}
		}
	}

	index.realFile.data = []byte(data)

	return index
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

// Close Implements the http.File interface
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
