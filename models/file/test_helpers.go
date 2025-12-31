package file

import (
	"time"

	file_system "github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/option"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestFileOptions provides options for creating test files.
type TestFileOptions struct {
	// Path is the portable path for the file (e.g., "USERS/testuser/file.txt")
	RootName string
	RelPath  string

	// ID is the file ID. If empty, a new ID will be generated.
	ID string

	// ContentID is the content hash for media files.
	ContentID string

	// IsDir indicates if this is a directory.
	IsDir bool

	// IsPastFile indicates this is a historical version.
	IsPastFile bool

	// Size is the file size in bytes.
	Size int64

	// ModTime is the modification time.
	ModTime time.Time

	// Children are the child files for directories.
	Children []*WeblensFileImpl

	// Parent is the parent directory.
	Parent *WeblensFileImpl

	// PastID is the ID reference for past files.
	PastID string
}

// NewTestFile creates a WeblensFileImpl for testing purposes.
// Uses MemOnly mode to avoid filesystem operations.
func NewTestFile(opts TestFileOptions) *WeblensFileImpl {
	relPath := opts.RelPath
	if opts.IsDir && len(relPath) > 0 && relPath[len(relPath)-1] != '/' {
		relPath += "/"
	}

	path := file_system.BuildFilePath(opts.RootName, relPath)

	id := opts.ID
	if id == "" {
		id = primitive.NewObjectID().Hex()
	}

	modTime := opts.ModTime
	if modTime.IsZero() {
		modTime = time.Now()
	}

	f := NewWeblensFile(NewFileOptions{
		Path:         path,
		FileID:       id,
		ContentID:    opts.ContentID,
		IsPastFile:   opts.IsPastFile,
		Size:         opts.Size,
		MemOnly:      true,
		ModifiedDate: option.Of(modTime),
	})

	// Set parent if provided
	if opts.Parent != nil {
		f.parent = opts.Parent
	}

	// Set children if provided
	if len(opts.Children) > 0 {
		f.childrenMap = make(map[string]*WeblensFileImpl)
		for _, child := range opts.Children {
			f.childrenMap[child.ID()] = child
			child.parent = f
		}
	}

	// Set past ID if provided
	if opts.PastID != "" {
		f.pastID = opts.PastID
	}

	return f
}

// NewTestDir creates a directory WeblensFileImpl for testing.
func NewTestDir(rootName, relPath string) *WeblensFileImpl {
	return NewTestFile(TestFileOptions{
		RootName: rootName,
		RelPath:  relPath,
		IsDir:    true,
	})
}
