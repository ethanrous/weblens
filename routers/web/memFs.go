package web

import (
	"bytes"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ethanrous/weblens/modules/config"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/services/context"
	"github.com/rs/zerolog/log"
)

// InMemoryFS implements an in-memory filesystem for serving web UI assets with template support.
type InMemoryFS struct {
	routes   map[string]*memFileReal
	index    *memFileReal
	routesMu *sync.RWMutex

	proxyAddress string
	uiPath       string
	ctx          context.AppContext
}

// Open opens the named file from the in-memory filesystem, implementing the http.FileSystem interface.
func (fs *InMemoryFS) Open(name string) (http.File, error) {
	log.Trace().Msgf("MemFs Opening file: %s", name)

	if name == "/index" {
		return nil, errors.New("index.html should be provided through the template")
	}

	var f *memFileReal

	var ok bool

	name = filepath.Join(config.GetConfig().UIPath, name)

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
		log.Error().Msgf("File %s does not exist", name)

		return nil, os.ErrNotExist
	}

	return newWrapFile(f), nil
}

func newWrapFile(r *memFileReal) *MemFileWrap {
	return &MemFileWrap{
		at:       0,
		realFile: r,
	}
}

// Exists checks if a file exists at the specified path in the in-memory filesystem.
func (fs *InMemoryFS) Exists(_ string, path string) bool {
	switch path {
	case "/", "//":
		return false
	case "/index":
		return true
	}

	_, err := fs.Open(path)

	return err == nil
}

// MemFileWrap wraps a memFileReal to provide http.File interface with read position tracking.
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
	URL         string
	Image       string
	Type        string
	VideoURL    string
	SecureURL   string
	VideoType   string
}

// Index returns the index.html file with template fields populated based on the request context.
func (fs InMemoryFS) Index(ctx context.RequestContext) *MemFileWrap {
	index := newWrapFile(fs.index.Copy())
	fields := getIndexFields(ctx, fs.proxyAddress)

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

func (fs *InMemoryFS) loadIndex(uiDir string) string {
	indexPath := filepath.Join(uiDir, "index.html")

	fs.index = readFile(indexPath, fs)
	if !fs.index.exists {
		panic(errors.Errorf("Could not find index file at %s", indexPath))
	}

	fs.uiPath = uiDir

	return indexPath
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

// Close closes the file, implementing the http.File interface.
func (f *MemFileWrap) Close() error {
	return nil
}

// Stat returns file information for the wrapped file, implementing the http.File interface.
func (f *MemFileWrap) Stat() (os.FileInfo, error) {
	return &InMemoryFileInfo{f.realFile}, nil
}

// Stat returns file information for the real file.
func (mf *memFileReal) Stat() (os.FileInfo, error) {
	return &InMemoryFileInfo{mf}, nil
}

// Readdir reads directory entries, implementing the http.File interface.
func (f *MemFileWrap) Readdir(_ int) ([]os.FileInfo, error) {
	res := make([]os.FileInfo, len(f.realFile.fs.routes))
	i := 0

	for _, file := range f.realFile.fs.routes {
		res[i], _ = file.Stat()
		i++
	}

	return res, nil
}

// Read reads file data into the provided byte slice, implementing the http.File interface.
func (f *MemFileWrap) Read(b []byte) (int, error) {
	n := copy(b, f.realFile.data[f.at:f.at+int64(len(b))])
	f.at += int64(len(b))

	return n, nil
}

// Seek sets the read position in the file, implementing the http.File interface.
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

// InMemoryFileInfo implements os.FileInfo for in-memory files.
type InMemoryFileInfo struct {
	file *memFileReal
}

// Name returns the file name, implementing os.FileInfo.
func (s *InMemoryFileInfo) Name() string { return s.file.Name }

// Size returns the file size in bytes, implementing os.FileInfo.
func (s *InMemoryFileInfo) Size() int64 { return int64(len(s.file.data)) }

// Mode returns the file mode, implementing os.FileInfo.
func (s *InMemoryFileInfo) Mode() os.FileMode { return os.ModeTemporary }

// ModTime returns the file modification time, implementing os.FileInfo.
func (s *InMemoryFileInfo) ModTime() time.Time { return time.Time{} }

// IsDir returns whether this is a directory, implementing os.FileInfo.
func (s *InMemoryFileInfo) IsDir() bool { return false }

// Sys returns the underlying data source, implementing os.FileInfo.
func (s *InMemoryFileInfo) Sys() any { return nil }
