package structs

type MediaInfo struct {
	MediaId string `json:"-" example:"5f9b3b3b7b4f3b0001b3b3b7"`

	// Hash of the file content, to ensure that the same files don't get duplicated
	ContentId string `json:"contentId"`

	// User who owns the file that resulted in this media being created
	Owner string `json:"owner"`

	// Mime-type key of the media
	MimeType string `json:"mimeType"`

	// Slices of files whos content hash to the contentId
	FileIds []string `json:"fileIds"`

	// Tags from the ML image scan so searching for particular objects in the images can be done
	RecognitionTags []string `json:"recognitionTags"`

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

	Imported bool `json:"imported"`
} // @Name MediaInfo

type MediaTypeInfo struct {
	MimeMap map[string]any `json:"mimeMap"`
	ExtMap  map[string]any `json:"extMap"`
} // @name MediaTypeInfo

type MediaBatchInfo struct {
	Media      []MediaInfo `json:"Media"`
	MediaCount int         `json:"mediaCount"`
} // @name MediaBatchInfo
