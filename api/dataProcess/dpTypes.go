package dataProcess

type SubscribeContent struct{
	Path string `json:"path"`
	Recursive bool `json:"recursive"`
}

type ScanContent struct{
	Path string `json:"path"`
	Recursive bool `json:"recursive"`
}

type WsMsg struct {
	Type string 	`json:"type"`
	Content any 	`json:"content"`
	Error string 	`json:"error"`
}