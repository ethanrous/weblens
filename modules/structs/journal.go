package structs

type FileActionInfo struct {
	FileId          string `json:"fileId" validate:"required"`
	ActionType      string `json:"actionType" validate:"required"`
	Filepath        string `json:"filepath,omitempty"`
	OriginPath      string `json:"originPath,omitempty"`
	DestinationPath string `json:"destinationPath,omitempty"`
	EventId         string `json:"eventId" validate:"required"`
	ParentId        string `json:"parentId" validate:"required"`
	TowerId         string `json:"towerId" validate:"required"`
	Timestamp       int64  `json:"timestamp" validate:"required" format:"int64"`
	Size            int64  `json:"size" validate:"required" format:"int64"`
} // @name FileActionInfo
