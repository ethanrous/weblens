// Package file provides file management functionality for the Weblens system.
package file

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"math"
	"time"

	"github.com/ethanrous/weblens/modules/option"
	"github.com/ethanrous/weblens/modules/wlerrors"
	file_system "github.com/ethanrous/weblens/modules/wlfs"
	"github.com/ethanrous/weblens/modules/wlog"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NewFileOptions represents configuration options for creating a new WeblensFile.
type NewFileOptions struct {
	Path file_system.Filepath

	FileID    string
	ContentID string

	IsPastFile bool

	Size int64

	MemOnly bool

	// Not directly used, but controls the creation of the file

	// CreateNow will create the file if it doesn't exist
	CreateNow bool

	// GenerateID will generate a new ID for the file
	GenerateID bool

	ModifiedDate option.Option[time.Time]
}

// NewWeblensFile creates and initializes a new WeblensFileImpl with the provided options.
func NewWeblensFile(params NewFileOptions) *WeblensFileImpl {
	// TODO: make this return an error instead of panicking
	if params.Path.IsZero() {
		panic("Path cannot be empty")
	}

	f := &WeblensFileImpl{
		portablePath: params.Path,
		isDir:        option.Of(params.Path.IsDir()),

		id:         params.FileID,
		contentID:  params.ContentID,
		pastFile:   params.IsPastFile,
		memOnly:    params.MemOnly,
		modifyDate: params.ModifiedDate.GetOr(time.Now()),
	}

	if params.Size > 0 {
		f.size.Store(params.Size)
	} else {
		f.size.Store(-1)
	}

	if params.CreateNow {
		err := f.CreateSelf()
		if wlerrors.Ignore(err, ErrFileAlreadyExists) != nil {
			wlog.GlobalLogger().Error().Err(err).Msgf("Failed to create file %s", params.Path.ToAbsolute())

			return nil
		}
	}

	if params.GenerateID {
		if params.Path.RelPath == "" {
			f.id = params.Path.RootName()
		} else {
			f.id = primitive.NewObjectID().Hex()
		}
	}

	return f
}

const oneMB = 1024 * 1024

// GenerateContentID computes and returns a content hash for the file.
func GenerateContentID(ctx context.Context, f *WeblensFileImpl) (string, error) {
	l := wlog.FromContext(ctx)

	if f.IsDir() {
		return "", wlerrors.Errorf("cannot hash directory")
	}

	if f.GetContentID() != "" {
		return f.GetContentID(), nil
	}

	l.Trace().Msgf("Generating content id for [%s]", f.GetPortablePath())

	fileSize := f.Size()

	if fileSize == 0 {
		return "", wlerrors.WithStack(ErrEmptyFile)
	}

	// Read up to 1MB at a time
	bufSize := math.Min(float64(fileSize), oneMB)
	buf := make([]byte, int64(bufSize))
	newHash := sha256.New()

	fp, err := f.Readable()
	if err != nil {
		return "", err
	}

	if closer, ok := fp.(io.Closer); ok {
		defer func(fp io.Closer) {
			_ = fp.Close()
		}(closer)
	}

	_, err = io.CopyBuffer(newHash, fp, buf)
	if err != nil {
		return "", err
	}

	contentID := base64.URLEncoding.EncodeToString(newHash.Sum(nil))[:20]
	f.SetContentID(contentID)

	return contentID, nil
}
