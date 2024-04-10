package routes

import (
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type InMemoryFS struct {
	routes   map[string]*memFileReal
	index    *memFileReal
	routesMu *sync.RWMutex
}

func (fs *InMemoryFS) loadIndex() string {
	indexPath, err := filepath.Abs("../ui/dist/index.html")
	if err != nil {
		panic("Could not find index file")
	}
	fs.index = ReadFile(indexPath, fs)
	if !fs.index.exists {
		panic("Could not find index file")
	}

	return indexPath
}

// Implements FileSystem interface
func (fs InMemoryFS) Open(name string) (http.File, error) {
	if name == "/index" {
		return newWrapFile(fs.index), nil
	}
	var f *memFileReal
	var ok bool
	name = filepath.Join("../ui/dist/", name)
	fs.routesMu.RLock()
	if f, ok = fs.routes[name]; ok && f.exists {
		fs.routesMu.RUnlock()
		return newWrapFile(f), nil
	} else if !ok {
		fs.routesMu.RUnlock()
		f = ReadFile(name, &fs)
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
		return newWrapFile(fs.index), nil
	}
	return newWrapFile(f), nil
}

func newWrapFile(real *memFileReal) *MemFileWrap {
	return &MemFileWrap{
		at:       0,
		realFile: real,
	}
}

func (f InMemoryFS) Exists(prefix string, path string) bool {
	if path == "/" || path == "//" {
		return false
	} else if path == "/index" {
		return true
	}
	_, err := f.Open(path)
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

func ReadFile(filePath string, fs *InMemoryFS) *memFileReal {
	fdata, err := os.ReadFile(filePath)

	return &memFileReal{
		Name:   filePath,
		data:   []byte(fdata),
		exists: err == nil,
		fs:     fs,
	}
}

// Implements the http.File interface
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

// Implements os.FileInfo
func (s *InMemoryFileInfo) Name() string       { return s.file.Name }
func (s *InMemoryFileInfo) Size() int64        { return int64(len(s.file.data)) }
func (s *InMemoryFileInfo) Mode() os.FileMode  { return os.ModeTemporary }
func (s *InMemoryFileInfo) ModTime() time.Time { return time.Time{} }
func (s *InMemoryFileInfo) IsDir() bool        { return false }
func (s *InMemoryFileInfo) Sys() interface{}   { return nil }