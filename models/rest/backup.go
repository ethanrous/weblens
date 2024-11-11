package rest

import (
	"encoding/json"

	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/models"
)

type BackupBody struct {
	FileHistory    []*fileTree.Lifetime
	LifetimesCount int
	Users          []*models.User
	Instances      []*models.Instance
	ApiKeys        []models.ApiKey
}

func (b BackupBody) MarshalJSON() ([]byte, error) {
	data := map[string]any{}

	var users []UserInfo
	for _, u := range b.Users {
		users = append(users, UserToUserInfo(u))
	}

	data["users"] = users
	data["fileHistory"] = b.FileHistory
	data["instances"] = b.Instances
	data["apiKeys"] = b.ApiKeys
	data["lifetimesCount"] = b.LifetimesCount

	return json.Marshal(data)
}
