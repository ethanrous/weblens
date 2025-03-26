package structs

type FileActionInfo struct {
	ActionType      string `json:"actionType" validate:"required"`
	OriginPath      string `json:"originPath" validate:"required"`
	DestinationPath string `json:"destinationPath" validate:"required"`
	LifeId          string `json:"lifeId" validate:"required"`
	EventId         string `json:"eventId" validate:"required"`
	ParentId        string `json:"parentId" validate:"required"`
	ServerId        string `json:"serverId" validate:"required"`
	Timestamp       int64  `json:"timestamp" validate:"required"`
	Size            int64  `json:"size" validate:"required"`
} // @name FileActionInfo
