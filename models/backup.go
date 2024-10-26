package models

import (
	"encoding/json"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal/log"
)

type BackupBody struct {
	FileHistory    []*fileTree.Lifetime
	LifetimesCount int
	Users          []*User
	Instances      []*Instance
	ApikKeys       []ApiKeyInfo
}

func (b BackupBody) MarshalJSON() ([]byte, error) {
	data := map[string]any{}
	users := []map[string]any{}
	for _, u := range b.Users {
		archive, err := u.FormatArchive()
		if err != nil {
			log.ErrTrace(err)
		}
		users = append(users, archive)
	}

	data["users"] = users
	data["fileHistory"] = b.FileHistory
	data["instances"] = b.Instances
	data["apiKeys"] = b.ApikKeys
	data["lifetimesCount"] = b.LifetimesCount

	return json.Marshal(data)
}
