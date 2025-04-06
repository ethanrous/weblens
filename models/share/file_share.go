package share

import (
	"context"
	"time"

	"github.com/ethanrous/weblens/models/db"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/slices"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const ShareCollectionKey = "shares"

var ErrShareNotFound = errors.New("share not found")
var ErrShareAlreadyExists = errors.New("share already exists")

type FileShare struct {
	// Accessors is a list of users that have access to the share
	Accessors []string  `bson:"accessors"`
	Enabled   bool      `bson:"enabled"`
	Expires   time.Time `bson:"expires"`
	FileId    string    `bson:"fileId"`
	Owner     string    `bson:"owner"`
	Public    bool      `bson:"public"`
	ShareId   string    `bson:"_id"`
	ShareName string    `bson:"shareName"`
	Updated   time.Time `bson:"updated"`
	Wormhole  bool      `bson:"wormhole"`
}

func NewFileShare(ctx context.Context, fileId string, owner *user_model.User, accessors []*user_model.User, public bool, wormhole bool) (*FileShare, error) {
	return &FileShare{
		ShareId: primitive.NewObjectID().Hex(),
		FileId:  fileId,
		Owner:   owner.GetUsername(),
		Accessors: slices.Map(
			accessors, func(u *user_model.User) string {
				return u.GetUsername()
			},
		),
		Public:   public,
		Wormhole: wormhole,
		Enabled:  true,
		Updated:  time.Now(),
	}, nil
}

func GetShareById(ctx context.Context, shareId string) (*FileShare, error) {
	collection, err := db.GetCollection(ctx, ShareCollectionKey)
	if err != nil {
		return nil, err
	}

	var share FileShare
	err = collection.FindOne(ctx, bson.M{"_id": shareId}).Decode(&share)
	if err != nil {
		return nil, errors.WithStack(ErrShareNotFound)
	}

	return &share, nil
}

func GetShareByFileId(ctx context.Context, fileId string) (*FileShare, error) {
	return nil, errors.New("not implemented")
}

func GetSharedWithUser(ctx context.Context, username string) ([]*FileShare, error) {
	return nil, nil
}

func DeleteShare(ctx context.Context, shareId string) error {
	return errors.New("not implemented")
}

func (s *FileShare) SetPublic(ctx context.Context, pub bool) error {
	return errors.New("not implemented")
}

func (s *FileShare) AddUsers(ctx context.Context, usernames []string) error {
	return errors.New("not implemented")
}

func (s *FileShare) RemoveUsers(ctx context.Context, usernames []string) error {
	return errors.New("not implemented")
}

func (s *FileShare) ID() string              { return s.ShareId }
func (s *FileShare) GetItemId() string       { return string(s.FileId) }
func (s *FileShare) SetItemId(fileId string) { s.FileId = fileId }
func (s *FileShare) GetAccessors() []string  { return s.Accessors }

func (s *FileShare) SetAccessors(usernames []string) {
	s.Accessors = usernames
}

func (s *FileShare) GetOwner() string { return s.Owner }
func (s *FileShare) IsPublic() bool   { return s.Public }
func (s *FileShare) IsWormhole() bool { return s.Wormhole }

func (s *FileShare) IsEnabled() bool { return s.Enabled }
func (s *FileShare) SetEnabled(enable bool) {
	s.Enabled = enable
}

func (s *FileShare) UpdatedNow() {
	s.Updated = time.Now()
}

func (s *FileShare) LastUpdated() time.Time {
	return s.Updated
}
