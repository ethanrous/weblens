package structs

// TakeoutInfo represents information about a data takeout export operation.
type TakeoutInfo struct {
	TakeoutID string `json:"takeoutID"`
	TaskID    string `json:"taskID"`
	Filename  string `json:"filename"`
	Single    bool   `json:"single"`
} // @name TakeoutInfo
