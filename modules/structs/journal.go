package structs

type FileActionInfo struct {
	ActionType      string `json:"actionType" validate:"required"`
	Filepath        string `json:"filepath,omitempty"`
	OriginPath      string `json:"originPath,omitempty"`
	DestinationPath string `json:"destinationPath,omitempty"`
	EventId         string `json:"eventId" validate:"required"`
	ParentId        string `json:"parentId" validate:"required"`
	TowerId         string `json:"towerId" validate:"required"`
	Timestamp       int64  `json:"timestamp" validate:"required"`
	Size            int64  `json:"size" validate:"required"`
} // @name FileActionInfo
