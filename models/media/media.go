package media

import (
	"context"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/ethanrous/weblens/models/db"
	file_model "github.com/ethanrous/weblens/models/file"
	slices_mod "github.com/ethanrous/weblens/modules/slices"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const MediaCollectionKey = "media"

var ErrMediaNotFound = errors.New("media not found")
var ErrMediaAlreadyExists = errors.New("media already exists")
var ErrNotDisplayable = errors.New("media is not displayable")
var ErrMediaBadMimeType = errors.New("media has a bad mime type")

type Media struct {
	CreateDate time.Time `bson:"createDate"`

	// WEBP thumbnail cache fileId
	lowresCacheFile *file_model.WeblensFileImpl

	// Hash of the file content, to ensure that the same files don't get duplicated
	ContentID ContentId `bson:"contentId"`

	// User who owns the file that resulted in this media being created
	Owner string `bson:"owner"`

	// Mime-type key of the media
	MimeType string `bson:"mimeType"`

	// The rotation of the image from its original. Found from the exif data
	Rotate string

	// Slices of files whos content hash to the contentId
	FileIDs []string `bson:"fileIds"`

	// Tags from the ML image scan so searching for particular objects in the images can be done
	RecognitionTags []string `bson:"recognitionTags"`

	LikedBy []string `bson:"likedBy"`

	// Ids for the files that are the cached WEBP of the fullres file. This is a slice
	// because fullres images could be multi-page, and a cache file is created per page
	highResCacheFiles []*file_model.WeblensFileImpl

	// Full-res image dimensions
	Width  int `bson:"width"`
	Height int `bson:"height"`

	// Number of pages (typically 1, 0 in not a valid page count)
	PageCount int `bson:"pageCount"`

	// Total time, in milliseconds, of a video
	Duration int `bson:"duration"`

	// Lock to synchronize updates to the media
	updateMu sync.RWMutex

	MediaID primitive.ObjectID `bson:"_id"`

	// If the media is hidden from the timeline
	// TODO - make this per user
	Hidden bool `bson:"hidden"`

	// If the media disabled. This can happen when the backing file(s) are deleted,
	// but the media stays behind because it can be re-used if needed.
	Enabled bool `bson:"enabled"`

	// If the media is imported into the databse yet. If not, we shouldn't ask about
	// things like cache, dimensions, etc., as it might not have them.
	imported bool
}

func GetMediaById(ctx context.Context, id ContentId) (*Media, error) {
	col, err := db.GetCollection(ctx, MediaCollectionKey)
	if err != nil {
		return nil, err
	}

	media := &Media{}
	err = col.FindOne(ctx, Media{ContentID: id}).Decode(&media)
	if err != nil {
		return nil, err
	}

	return media, nil
}

func GetMediaByPath(ctx context.Context, path string) ([]*Media, error) {
	// col, err := db.GetCollection(ctx, MediaCollectionKey)
	// if err != nil {
	// 	return nil, err
	// }
	//
	// media := &Media{}
	// ret, err := col.Find(ctx, Media{})
	// if err != nil {
	// 	return nil, err
	// }
	//
	// return media, nil
	return nil, errors.New("not implemented")
}

func GetMedia(ctx context.Context, username string, sort string, sortDirection int, excludeIds []ContentId,
	allowRaw bool, allowHidden bool, search string) ([]*Media, error) {

	slices.Sort(excludeIds)

	pipe := bson.A{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "owner", Value: username},
				{Key: "fileIds", Value: bson.D{
					{Key: "$exists", Value: true}, {Key: "$ne", Value: bson.A{}},
				}}},
			},
		},
	}

	if !allowHidden {
		pipe = append(pipe, bson.D{{Key: "$match", Value: bson.D{{Key: "hidden", Value: false}}}})
	}

	if search != "" {
		search = strings.ToLower(search)
		pipe = append(pipe, bson.D{{Key: "$match", Value: bson.D{{Key: "recognitionTags", Value: bson.D{{Key: "$regex", Value: search}}}}}})
	}

	pipe = append(pipe, bson.D{{Key: "$sort", Value: bson.D{{Key: sort, Value: sortDirection}}}})
	// pipe = append(pipe, bson.D{{Key: "$project", Value: bson.D{{Key: "_id", Value: false}, {Key: "contentId", Value: true}}}})

	col, err := db.GetCollection(ctx, MediaCollectionKey)
	if err != nil {
		return nil, err
	}

	cur, err := col.Aggregate(ctx, pipe)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	medias := []*Media{}
	err = cur.All(ctx, &medias)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return medias, nil
}

func DropMediaCollection(ctx context.Context) error {
	col, err := db.GetCollection(ctx, MediaCollectionKey)
	if err != nil {
		return err
	}

	err = col.Drop(ctx)
	if err != nil {
		return err
	}

	return nil
}

func NewMedia(contentId ContentId) *Media {
	return &Media{
		ContentID: contentId,
	}
}

func (m *Media) ID() ContentId {
	return m.ContentID
}

func (m *Media) SetContentId(id ContentId) {
	m.ContentID = id
}

func (m *Media) IsHidden() bool {
	return m.Hidden
}

func (m *Media) GetCreateDate() time.Time {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()
	return m.CreateDate
}

func (m *Media) SetCreateDate(t time.Time) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()
	m.CreateDate = t
}

func (m *Media) GetPageCount() int {
	return m.PageCount
}

// GetVideoLength returns the length of the media video, if it is a video. Duration is counted in milliseconds
func (m *Media) GetVideoLength() int {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()
	return m.Duration
}

func (m *Media) SetOwner(owner string) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()
	m.Owner = owner
}

func (m *Media) GetOwner() string {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()
	return m.Owner
}

func (m *Media) GetFiles() []string {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()
	return m.FileIDs
}

func (m *Media) AddFile(f *file_model.WeblensFileImpl) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()
	m.FileIDs = slices_mod.AddToSet(m.FileIDs, f.ID())
}

func (m *Media) RemoveFile(fileIdToRemove string) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()
	m.FileIDs = slices_mod.Filter(
		m.FileIDs, func(fId string) bool {
			return fId != fileIdToRemove
		},
	)
}

func (m *Media) SetImported(i bool) {
	m.imported = i
}

func (m *Media) IsImported() bool {
	if m == nil {
		return false
	}

	return m.imported
}

func (m *Media) SetEnabled(e bool) {
	m.Enabled = e
}

func (m *Media) IsEnabled() bool {
	return m.Enabled
}

func (m *Media) SetRecognitionTags(tags []string) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()
	m.RecognitionTags = tags
}

func (m *Media) GetRecognitionTags() []string {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()
	return m.RecognitionTags
}

func (m *Media) setHidden(hidden bool) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()
	m.Hidden = hidden
}

func (m *Media) SetLowresCacheFile(thumb *file_model.WeblensFileImpl) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()
	m.lowresCacheFile = thumb
}

func (m *Media) GetLowresCacheFile() *file_model.WeblensFileImpl {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()
	return m.lowresCacheFile
}

func (m *Media) SetHighresCacheFiles(highresFile *file_model.WeblensFileImpl, pageNum int) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()
	if len(m.highResCacheFiles) < pageNum+1 {
		m.highResCacheFiles = make([]*file_model.WeblensFileImpl, m.PageCount)
	}
	m.highResCacheFiles[pageNum] = highresFile
}

func (m *Media) GetHighresCacheFiles(pageNum int) *file_model.WeblensFileImpl {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()
	if len(m.highResCacheFiles) < pageNum+1 {
		return nil
	}
	return m.highResCacheFiles[pageNum]
}

const ThumbnailHeight float32 = 500

// func (m *Media) UnmarshalBSON(bs []byte) error {
// 	raw := bson.Raw(bs)
// 	contentId, ok := raw.Lookup("contentId").StringValueOK()
// 	if !ok {
// 		return werror.Errorf("failed to parse contentId")
// 	}
// 	m.ContentID = ContentId(contentId)
//
// 	filesArr, ok := raw.Lookup("fileIds").ArrayOK()
// 	if ok {
// 		fileIdsRaw, err := filesArr.Values()
// 		if err != nil {
// 			return werror.WithStack(err)
// 		}
// 		m.FileIDs = slices.Map(
// 			fileIdsRaw, func(e bson.RawValue) string {
// 				return string(e.StringValue())
// 			},
// 		)
// 	} else {
// 		m.FileIDs = []string{}
// 	}
//
// 	m.PageCount = int(raw.Lookup("pageCount").Int32())
//
// 	videoLength, ok := raw.Lookup("videoLength").Int32OK()
// 	if ok {
// 		m.Duration = int(videoLength)
// 	}
//
// 	// m.BlurHash = raw.Lookup("blurHash").StringValue()
// 	m.Owner = raw.Lookup("owner").StringValue()
// 	m.Width = int(raw.Lookup("width").Int32())
// 	m.Height = int(raw.Lookup("height").Int32())
//
// 	duration, ok := raw.Lookup("duration").Int32OK()
// 	if ok {
// 		m.Duration = int(duration)
// 	}
//
// 	create := raw.Lookup("createDate")
// 	createTime, ok := create.TimeOK()
// 	if !ok {
// 		m.CreateDate = time.UnixMilli(create.Int64())
// 	} else {
// 		m.CreateDate = createTime
// 	}
// 	m.MimeType = raw.Lookup("mimeType").StringValue()
//
// 	likedArr, ok := raw.Lookup("likedBy").ArrayOK()
// 	if ok {
// 		likedValues, err := likedArr.Values()
// 		if err != nil {
// 			return werror.WithStack(err)
// 		}
// 		m.LikedBy = internal.Map(
// 			likedValues, func(e bson.RawValue) Username {
// 				return Username(e.StringValue())
// 			},
// 		)
// 	} else {
// 		m.LikedBy = []Username{}
// 	}
//
// 	rtArr, ok := raw.Lookup("recognitionTags").ArrayOK()
// 	if ok {
// 		rts, err := rtArr.Values()
// 		if err != nil {
// 			return werror.WithStack(err)
// 		}
//
// 		m.SetRecognitionTags(internal.Map(
// 			rts, func(e bson.RawValue) string {
// 				return e.StringValue()
// 			},
// 		))
// 	}
//
// 	hidden, ok := raw.Lookup("hidden").BooleanOK()
// 	if ok {
// 		m.setHidden(hidden)
// 	}
//
// 	m.imported = true
//
// 	return nil
// }

// type MediaService interface {
// 	Size() int
// 	Add(media *Media) error
//
// 	Get(id ContentId) *Media
// 	GetAll() []*Media
//
// 	Del(id ContentId) error
// 	HideMedia(m *Media, hidden bool) error
// 	AdjustMediaDates(anchor *Media, newTime time.Time, extraMedias []*Media) error
//
// 	LoadMediaFromFile(m *Media, file *file_model.WeblensFileImpl) error
// 	RemoveFileFromMedia(media *Media, fileId string) error
// 	Cleanup() error
// 	Drop() error
// 	AddFileToMedia(media *Media, file *file_model.WeblensFileImpl) error
//
// 	GetMediaType(m *Media) MediaType
// 	GetMediaTypes() MediaTypeService
// 	IsFileDisplayable(file *file_model.WeblensFileImpl) bool
// 	IsCached(m *Media) bool
// 	GetProminentColors(media *Media) (prom []string, err error)
// 	GetMediaConverted(m *Media, format string) ([]byte, error)
//
// 	FetchCacheImg(m *Media, quality MediaQuality, pageNum int) ([]byte, error)
// 	StreamVideo(m *Media, u *user_model.User, share *share_model.FileShare) (*VideoStreamer, error)
// 	StreamCacheVideo(m *Media, startByte, endByte int) ([]byte, error)
//
// 	GetFilteredMedia(
// 		requester *user_model.User, sort string, sortDirection int, excludeIds []ContentId, raw bool, hidden bool, search string,
// 	) ([]*Media, error)
// 	RecursiveGetMedia(folders ...*file_model.WeblensFileImpl) []*Media
//
// 	SetMediaLiked(mediaId ContentId, liked bool, username Username) error
// }

type ContentId = string
type MediaQuality string

const (
	LowRes  MediaQuality = "thumbnail"
	HighRes MediaQuality = "fullres"
	Video   MediaQuality = "video"
)

func CheckMediaQuality(quality string) (MediaQuality, bool) {
	switch quality {
	case string(LowRes), string(HighRes), string(Video):
		return MediaQuality(quality), true
	}
	return "", false
}

func RemoveFileFromMedia(ctx context.Context, media *Media, fileId string) error {
	col, err := db.GetCollection(ctx, MediaCollectionKey)
	if err != nil {
		return err
	}

	_, err = col.UpdateOne(ctx, bson.D{
		{Key: "contentId", Value: media.ContentID},
	}, bson.D{{Key: "$pull", Value: bson.D{{Key: "fileIds", Value: fileId}}}})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (media *Media) AddFileToMedia(ctx context.Context, fileId string) error {
	col, err := db.GetCollection(ctx, MediaCollectionKey)
	if err != nil {
		return err
	}

	_, err = col.UpdateOne(ctx, bson.D{
		{Key: "contentId", Value: media.ContentID},
	}, bson.D{{Key: "$addToSet", Value: bson.D{{Key: "fileIds", Value: fileId}}}})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
