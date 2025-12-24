package file

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"math"
	"time"

	"github.com/ethanrous/weblens/modules/errors"
	file_system "github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/modules/option"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NewFileOptions struct {
	Path file_system.Filepath

	FileId    string
	ContentId string

	IsPastFile bool

	Size int64

	MemOnly bool

	// Not directly used, but controls the creation of the file

	// CreateNow will create the file if it doesn't exist
	CreateNow bool

	// GenerateId will generate a new ID for the file
	GenerateId bool

	ModifiedDate option.Option[time.Time]
}

func NewWeblensFile(params NewFileOptions) *WeblensFileImpl {
	// TODO: make this return an error instead of panicking
	if params.Path.IsZero() {
		panic("Path cannot be empty")
	}

	f := &WeblensFileImpl{
		portablePath: params.Path,
		isDir:        option.Of(params.Path.IsDir()),

		id:         params.FileId,
		contentId:  params.ContentId,
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
		if errors.Ignore(err, ErrFileAlreadyExists) != nil {
			log.GlobalLogger().Error().Err(err).Msgf("Failed to create file %s", params.Path.ToAbsolute())

			return nil
		}
	}

	if params.GenerateId {
		if params.Path.RelPath == "" {
			f.id = params.Path.RootName()
		} else {
			f.id = primitive.NewObjectID().Hex()
		}
	}

	return f
}

const oneMB = 1024 * 1024

func GenerateContentId(ctx context.Context, f *WeblensFileImpl) (string, error) {
	l := log.FromContext(ctx)

	if f.IsDir() {
		return "", errors.Errorf("cannot hash directory")
	}

	if f.GetContentId() != "" {
		return f.GetContentId(), nil
	}

	l.Trace().Msgf("Generating content id for [%s]", f.GetPortablePath())

	fileSize := f.Size()

	if fileSize == 0 {
		return "", errors.WithStack(ErrEmptyFile)
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

	contentId := base64.URLEncoding.EncodeToString(newHash.Sum(nil))[:20]
	f.SetContentId(contentId)

	return contentId, nil
}
