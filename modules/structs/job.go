package structs

type TakeoutInfo struct {
	TakeoutId string `json:"takeoutId"`
	TaskId    string `json:"taskId"`
	Filename  string `json:"filename"`
	Single    bool   `json:"single"`
} // @name TakeoutInfo
