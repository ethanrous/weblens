package structs

import "time"

type Token struct {
	Id          string    `json:"id"`
	CreatedTime time.Time `json:"createdTime"`
	LastUsed    time.Time `json:"lastUsed"`
	Nickname    string    `json:"nickname"`
	Owner       string    `json:"owner"`
	RemoteUsing string    `json:"remoteUsing"`
	CreatedBy   string    `json:"createdBy"`
	Token       [32]byte  `json:"token"`
}
