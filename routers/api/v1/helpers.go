package v1

import (
	"time"
)

type FileStat struct {
	ModTime time.Time `json:"modifyTimestamp"`
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	IsDir   bool      `json:"isDir"`
	Exists  bool      `json:"exists"`
}
