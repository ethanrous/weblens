package media

import (
	"context"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/ethanrous/weblens/models/db"
	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/modules/errors"
	slices_mod "github.com/ethanrous/weblens/modules/slices"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MediaCollectionKey is the MongoDB collection name for media documents.
const MediaCollectionKey = "media"

// ErrMediaNotFound is returned when a requested media item does not exist.
var ErrMediaNotFound = errors.New("media not found")

// ErrMediaAlreadyExists is returned when attempting to create a media that already exists.
var ErrMediaAlreadyExists = errors.New("media already exists")

// ErrNotDisplayable is returned when media cannot be displayed.
var ErrNotDisplayable = errors.New("media is not displayable")

// ErrMediaBadMimeType is returned when media has an unsupported mime type.
var ErrMediaBadMimeType = errors.New("media has a bad mime type")

// ErrInvalidQuality is returned when an invalid media quality is specified.
var ErrInvalidQuality = errors.Errorf("invalid media quality")

// Media represents a media item stored in the database.
type Media struct {
	CreateDate time.Time `bson:"createDate"`

	// WEBP thumbnail cache fileId
	lowresCacheFile *file_model.WeblensFileImpl

	// Hash of the file content, to ensure that the same files don't get duplicated
	ContentID ContentID `bson:"contentID"`

	// User who owns the file that resulted in this media being created
	Owner string `bson:"owner"`

	// Mime-type key of the media
	MimeType string `bson:"mimeType"`

	// The rotation of the image from its original. Found from the exif data
	Rotate string

	// Location of the media, if available. This is a slice of two floats, representing the coordinates of the media.
	Location [2]float64 `bson:"location"`

	// High Dimensional Image Representation, a slice of floats representing the image's high-dimensional representation.
	// Used for image similarity searches.
	HDIR []float64 `bson:"hdir"`

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
	// TODO: make this per user
	Hidden bool `bson:"hidden"`

	// If the media disabled. This can happen when the backing file(s) are deleted,
	// but the media stays behind because it can be re-used if needed.
	Enabled bool `bson:"enabled"`

	// If the media is imported into the databse yet. If not, we shouldn't ask about
	// things like cache, dimensions, etc., as it might not have them.
	imported bool
}

// SaveMedia persists a media item to the database, upserting by content ID.
func SaveMedia(ctx context.Context, media *Media) error {
	if media.MediaID.IsZero() {
		media.MediaID = primitive.NewObjectID()
	}

	col, err := db.GetCollection[any](ctx, MediaCollectionKey)
	if err != nil {
		return err
	}

	_, err = col.ReplaceOne(ctx, bson.M{"contentID": media.ContentID}, media, options.Replace().SetUpsert(true))
	if err != nil {
		return db.WrapError(err, "insert media")
	}

	return nil
}

// GetMediaByContentID retrieves a media item by its content ID.
func GetMediaByContentID(ctx context.Context, contentID ContentID) (*Media, error) {
	col, err := db.GetCollection[any](ctx, MediaCollectionKey)
	if err != nil {
		return nil, err
	}

	media := Media{}

	err = col.FindOne(ctx, bson.M{"contentID": contentID}).Decode(&media)
	if err != nil {
		return nil, db.WrapError(err, "get media by content id")
	}

	return &media, nil
}

// GetMediasByContentIDs retrieves multiple media items by their content IDs with pagination.
func GetMediasByContentIDs(ctx context.Context, limit, page, sortDirection int, includeRaw bool, contentIDs ...ContentID) ([]*Media, error) {
	col, err := db.GetCollection[any](ctx, MediaCollectionKey)
	if err != nil {
		return nil, err
	}

	media := []*Media{}

	filter := bson.M{"contentID": bson.M{"$in": contentIDs}, "duration": bson.M{"$eq": 0}}
	if !includeRaw {
		filter["mimeType"] = bson.M{"$not": bson.M{"$in": rawMimes()}}
	}

	cur, err := col.Find(ctx, filter, options.Find().SetLimit(int64(limit)).SetSkip(int64(page*limit)).SetSort(bson.D{{Key: "createDate", Value: sortDirection}}))
	if err != nil {
		return nil, err
	}

	err = cur.All(ctx, &media)
	if err != nil {
		return nil, db.WrapError(err, "get media by contentIds")
	}

	return media, nil
}

// GetMediaByPath retrieves media items by file path.
func GetMediaByPath(_ context.Context, _ string) ([]*Media, error) {
	// col, err := db.GetCollection[any](ctx, MediaCollectionKey)
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

// GetMedia retrieves media items for a user with filtering and sorting options.
func GetMedia(ctx context.Context, username string, sort string, sortDirection int, excludeIDs []ContentID,
	_ bool, allowHidden bool, search string) ([]*Media, error) {
	slices.Sort(excludeIDs)

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
	// pipe = append(pipe, bson.D{{Key: "$project", Value: bson.D{{Key: "_id", Value: false}, {Key: "contentID", Value: true}}}})

	col, err := db.GetCollection[any](ctx, MediaCollectionKey)
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

// RandomMediaOptions specifies options for retrieving random media items.
type RandomMediaOptions struct {
	Count  int
	Owner  string
	NoRaws bool // If true, do not return raw media files
}

// GetRandomMedias retrieves a random sample of media items based on the given options.
func GetRandomMedias(ctx context.Context, opts RandomMediaOptions) ([]*Media, error) {
	col, err := db.GetCollection[any](ctx, MediaCollectionKey)
	if err != nil {
		return nil, err
	}

	match := bson.M{
		// Media must have a file associated with it
		"fileIds": bson.M{"$exists": true, "$ne": bson.A{}},
		"hidden":  false, // Only return non-hidden media
	}

	if opts.Owner != "" {
		// Match the owner if given
		match["owner"] = opts.Owner
	}

	if opts.NoRaws {
		// Match the owner if given
		match["mimeType"] = bson.M{"$ne": "image/x-sony-arw"}
	}

	cursor, err := col.Aggregate(ctx, bson.A{
		bson.M{"$match": match},
		// Sample the given number of random documents
		bson.M{"$sample": bson.M{"size": opts.Count}},
	})
	if err != nil {
		return nil, err
	}

	var target []*Media

	err = cursor.All(ctx, &target)
	if err != nil {
		return nil, err
	}

	return target, nil
}

// DropMediaCollection removes the entire media collection from the database.
func DropMediaCollection(ctx context.Context) error {
	col, err := db.GetCollection[any](ctx, MediaCollectionKey)
	if err != nil {
		return err
	}

	err = col.Drop(ctx)
	if err != nil {
		return err
	}

	return nil
}

// DropHDIRs removes all HDIR (High Dimensional Image Representation) data from media documents.
func DropHDIRs(ctx context.Context) error {
	col, err := db.GetCollection[any](ctx, MediaCollectionKey)
	if err != nil {
		return err
	}

	_, err = col.UpdateMany(ctx, bson.M{}, bson.M{"$unset": bson.M{"hdir": ""}})
	if err != nil {
		return err
	}

	return nil
}

// NewMedia creates a new Media instance with the given content ID.
func NewMedia(contentID ContentID) *Media {
	return &Media{
		ContentID: contentID,
	}
}

// ID returns the content ID of the media.
func (m *Media) ID() ContentID {
	return m.ContentID
}

// SetContentID sets the content ID of the media.
func (m *Media) SetContentID(id ContentID) {
	m.ContentID = id
}

// IsHidden returns true if the media is marked as hidden.
func (m *Media) IsHidden() bool {
	return m.Hidden
}

// GetCreateDate returns the creation date of the media.
func (m *Media) GetCreateDate() time.Time {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()

	return m.CreateDate
}

// SetCreateDate sets the creation date of the media.
func (m *Media) SetCreateDate(t time.Time) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()

	m.CreateDate = t
}

// GetPageCount returns the number of pages in the media.
func (m *Media) GetPageCount() int {
	return m.PageCount
}

// GetVideoLength returns the length of the media video, if it is a video. Duration is counted in milliseconds
func (m *Media) GetVideoLength() int {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()

	return m.Duration
}

// SetOwner sets the owner username for the media.
func (m *Media) SetOwner(owner string) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()

	m.Owner = owner
}

// GetOwner returns the owner username of the media.
func (m *Media) GetOwner() string {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()

	return m.Owner
}

// GetFiles returns the list of file IDs associated with the media.
func (m *Media) GetFiles() []string {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()

	return m.FileIDs
}

// AddFile adds a file ID to the media's file list.
func (m *Media) AddFile(f *file_model.WeblensFileImpl) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()

	m.FileIDs = slices_mod.AddToSet(m.FileIDs, f.ID())
}

// RemoveFile removes a file from this media's list of associated files.
func (m *Media) RemoveFile(fileIDToRemove string) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()

	m.FileIDs = slices_mod.Filter(
		m.FileIDs, func(fID string) bool {
			return fID != fileIDToRemove
		},
	)
}

// SetImported sets whether the media has been imported.
func (m *Media) SetImported(i bool) {
	m.imported = i
}

// IsImported returns true if the media has been imported.
func (m *Media) IsImported() bool {
	if m == nil {
		return false
	}

	return m.imported
}

// SetEnabled sets whether the media is enabled.
func (m *Media) SetEnabled(e bool) {
	m.Enabled = e
}

// IsEnabled returns true if the media is enabled.
func (m *Media) IsEnabled() bool {
	return m.Enabled
}

// SetRecognitionTags sets the object recognition tags for the media.
func (m *Media) SetRecognitionTags(tags []string) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()

	m.RecognitionTags = tags
}

// GetRecognitionTags returns the object recognition tags for the media.
func (m *Media) GetRecognitionTags() []string {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()

	return m.RecognitionTags
}

// SetLowresCacheFile sets the thumbnail/low-resolution cache file.
func (m *Media) SetLowresCacheFile(thumb *file_model.WeblensFileImpl) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()

	m.lowresCacheFile = thumb
}

// GetLowresCacheFile returns the thumbnail/low-resolution cache file.
func (m *Media) GetLowresCacheFile() *file_model.WeblensFileImpl {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()

	return m.lowresCacheFile
}

// SetHighresCacheFiles sets the high-resolution cache file for a specific page.
func (m *Media) SetHighresCacheFiles(highresFile *file_model.WeblensFileImpl, pageNum int) {
	m.updateMu.Lock()
	defer m.updateMu.Unlock()

	if len(m.highResCacheFiles) < pageNum+1 {
		m.highResCacheFiles = make([]*file_model.WeblensFileImpl, m.PageCount)
	}

	m.highResCacheFiles[pageNum] = highresFile
}

// GetHighresCacheFiles returns the high-resolution cache file for a specific page.
func (m *Media) GetHighresCacheFiles(pageNum int) *file_model.WeblensFileImpl {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()

	if len(m.highResCacheFiles) < pageNum+1 {
		return nil
	}

	return m.highResCacheFiles[pageNum]
}

// IsSufficentlyProcessed returns true if the media has been sufficiently processed.
func (m *Media) IsSufficentlyProcessed(requireHDIR bool) bool {
	m.updateMu.RLock()
	defer m.updateMu.RUnlock()

	if len(m.FileIDs) == 0 {
		return false
	}

	if requireHDIR && len(m.HDIR) == 0 {
		return false
	}

	return true
}

// ThumbnailHeight is the standard height for thumbnail images.
const ThumbnailHeight float32 = 500

// ContentID is a unique identifier for media content.
type ContentID = string

// Quality represents the quality level of media.
type Quality string

// Media quality constants.
const (
	LowRes  Quality = "thumbnail"
	HighRes Quality = "fullres"
	Video   Quality = "video"
)

// CheckMediaQuality validates and converts a quality string to MediaQuality.
func CheckMediaQuality(quality string) (Quality, bool) {
	switch quality {
	case string(LowRes), string(HighRes), string(Video):
		return Quality(quality), true
	}

	return "", false
}

// RemoveFileFromMedia removes a file from a media's list of associated files in the database.
func RemoveFileFromMedia(ctx context.Context, media *Media, fileID string) error {
	col, err := db.GetCollection[any](ctx, MediaCollectionKey)
	if err != nil {
		return err
	}

	_, err = col.UpdateOne(ctx, bson.D{
		{Key: "contentID", Value: media.ContentID},
	}, bson.D{{Key: "$pull", Value: bson.D{{Key: "fileIds", Value: fileID}}}})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// AddFileToMedia adds a file to this media's list of associated files in the database.
func (m *Media) AddFileToMedia(ctx context.Context, fileID string) error {
	col, err := db.GetCollection[any](ctx, MediaCollectionKey)
	if err != nil {
		return err
	}

	_, err = col.UpdateOne(ctx, bson.D{
		{Key: "contentID", Value: m.ContentID},
	}, bson.D{{Key: "$addToSet", Value: bson.D{{Key: "fileIds", Value: fileID}}}})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
