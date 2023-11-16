package dataProcess

// Websocket request types
type WsRequest struct {
	ReqType string 	`json:"req"`
	Content any 	`json:"content"`
	Error string 	`json:"error"`
}

type SubscribeReqContent struct{
	SubType string `json:"type"`
	Metadata string `json:"metadata"`
}

type PathSubMetadata struct {
	DirPath string `json:"dirPath"`
	Recursive bool `json:"recursive"`
}

type TaskSubMetadata struct {
	TaskId string 		`json:"taskId"`
	LookingFor []string `json:"lookingFor"`
}

type ScanContent struct{
	Path string `json:"dirPath"`
	Recursive bool `json:"recursive"`
}

// Internal types
type ScanMetadata struct {
	Path string
	Username string
	Recursive bool
}

type ZipMetadata struct {
	Paths []string
	Username string
}

// Ws response types
type WsResponse struct {
	MessageStatus string 	`json:"messageStatus"`
	Content any 			`json:"content"`
	Error error 			`json:"error"`
}

// Misc
type KeyVal struct {
	Key string
	Val string
}