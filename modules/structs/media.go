package structs

// MediaInfo represents metadata and properties of a media file.
type MediaInfo struct {
	MediaID string `json:"-" example:"5f9b3b3b7b4f3b0001b3b3b7"`

	// Hash of the file content, to ensure that the same files don't get duplicated
	ContentID string `json:"contentID"`

	// User who owns the file that resulted in this media being created
	Owner string `json:"owner"`

	// Mime-type key of the media
	MimeType string `json:"mimeType"`

	Location [2]float64 `json:"location"`

	// Slices of files whos content hash to the contentId
	FileIDs []string `json:"fileIDs"`

	LikedBy []string `json:"likedBy,omitempty"`

	CreateDate int64 `json:"createDate"`

	// Full-res image dimensions
	Width  int `json:"width"`
	Height int `json:"height"`

	// Number of pages (typically 1, 0 in not a valid page count)
	PageCount int `json:"pageCount"`

	// Total time, in milliseconds, of a video
	Duration int `json:"duration"`

	// If the media is hidden from the timeline
	// TODO - make this per user
	Hidden bool `json:"hidden"`

	// If the media disabled. This can happen when the backing file(s) are deleted,
	// but the media stays behind because it can be re-used if needed.
	Enabled bool `json:"enabled"`

	// Similarity score from HDIR search
	HDIRScore float64 `json:"hdirScore,omitempty"`

	Imported bool `json:"imported"`
} // @Name MediaInfo

// MediaTypeInfo represents information about a specific media type.
type MediaTypeInfo struct {
	Mime            string   `json:"mime"`
	Name            string   `json:"FriendlyName"`
	RawThumbExifKey string   `json:"RawThumbExifKey"`
	Extensions      []string `json:"FileExtension"`
	Displayable     bool     `json:"IsDisplayable"`
	Raw             bool     `json:"IsRaw"`
	Video           bool     `json:"IsVideo"`
	ImgRecog        bool     `json:"SupportsImgRecog"`
	MultiPage       bool     `json:"MultiPage"`
} // @name MediaTypeInfo

// MediaTypesInfo represents the complete mapping of media types indexed by both mime type and file extension.
type MediaTypesInfo struct {
	MimeMap map[string]MediaTypeInfo `json:"mimeMap"`
	ExtMap  map[string]MediaTypeInfo `json:"extMap"`
} // @name MediaTypesInfo

// MediaBatchInfo represents a paginated batch of media items with count information.
type MediaBatchInfo struct {
	Media           []MediaInfo `json:"Media"`
	MediaCount      int         `json:"mediaCount"`
	TotalMediaCount int         `json:"totalMediaCount"`
} // @name MediaBatchInfo
