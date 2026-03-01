package wlstructs

// FileActionInfo represents a file system action event in the journal.
type FileActionInfo struct {
	FileID          string `json:"fileID" validate:"required"`
	ActionType      string `json:"actionType" validate:"required"`
	Filepath        string `json:"filepath,omitempty"`
	OriginPath      string `json:"originPath,omitempty"`
	DestinationPath string `json:"destinationPath,omitempty"`
	EventID         string `json:"eventID" validate:"required"`
	ParentID        string `json:"parentID" validate:"required"`
	TowerID         string `json:"towerID" validate:"required"`
	Timestamp       int64  `json:"timestamp" validate:"required" format:"int64"`
	Size            int64  `json:"size" validate:"required" format:"int64"`
	ContentID       string `json:"contentID,omitempty"`
} // @name FileActionInfo
