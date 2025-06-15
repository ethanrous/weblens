package share

import (
	"context"
	"time"

	"github.com/ethanrous/weblens/models/db"
	user_model "github.com/ethanrous/weblens/models/user"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/modules/slices"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const ShareCollectionKey = "shares"

var ErrShareNotFound = errors.New("share not found")
var ErrShareAlreadyExists = errors.New("share already exists")

type FileShare struct {
	// Accessors is a list of users that have access to the share
	Accessors []string `bson:"accessors"`
	// Permissions maps usernames to their specific permissions
	Permissions map[string]*Permissions `bson:"permissions"`
	Enabled     bool                    `bson:"enabled"`
	Expires     time.Time               `bson:"expires"`
	FileId      string                  `bson:"fileId"`
	Owner       string                  `bson:"owner"`
	Public      bool                    `bson:"public"`
	ShareId     primitive.ObjectID      `bson:"_id"`
	ShareName   string                  `bson:"shareName"`
	Updated     time.Time               `bson:"updated"`
	Wormhole    bool                    `bson:"wormhole"`
}

var IndexModels = []mongo.IndexModel{
	{
		Keys: bson.M{
			"fileId": -1,
		},
		Options: options.Index().SetUnique(true),
	},
}

func ShareIdFromString(shareId string) primitive.ObjectID {
	if shareId == "" {
		return primitive.NilObjectID
	}

	id, err := primitive.ObjectIDFromHex(shareId)
	if err != nil {
		log.FromContext(context.TODO()).Error().Err(err).Msgf("failed to parse share ID [%s]", shareId)

		return primitive.NilObjectID
	}

	return id
}

func NewFileShare(ctx context.Context, fileId string, owner *user_model.User, accessors []*user_model.User, public bool, wormhole bool) (*FileShare, error) {
	permissions := make(map[string]*Permissions)
	for _, u := range accessors {
		permissions[u.GetUsername()] = NewPermissions() // default permissions
	}

	return &FileShare{
		FileId: fileId,
		Owner:  owner.GetUsername(),
		Accessors: slices.Map(
			accessors, func(u *user_model.User) string {
				return u.GetUsername()
			},
		),
		Permissions: permissions,
		Public:      public,
		Wormhole:    wormhole,
		Enabled:     true,
		Updated:     time.Now(),
	}, nil
}

func SaveFileShare(ctx context.Context, share *FileShare) error {
	collection, err := db.GetCollection(ctx, ShareCollectionKey)
	if err != nil {
		return err
	}

	if share.ShareId.IsZero() {
		share.ShareId = primitive.NewObjectID()
	}

	if share.Accessors == nil {
		share.Accessors = []string{}
	}

	if share.Permissions == nil {
		share.Permissions = make(map[string]*Permissions)
	}

	_, err = collection.InsertOne(ctx, share)
	if err != nil {
		return db.WrapError(err, "failed to save share [%s]", share.ShareId)
	}

	return nil
}

func GetShareById(ctx context.Context, shareId primitive.ObjectID) (*FileShare, error) {
	collection, err := db.GetCollection(ctx, ShareCollectionKey)
	if err != nil {
		return nil, err
	}

	share := &FileShare{}

	err = collection.FindOne(ctx, bson.M{"_id": shareId}).Decode(share)
	if err != nil {
		return nil, db.WrapError(errors.WithStack(err), "failed to get share by id [%s]", shareId)
	}

	return share, nil
}

func GetShareByFileId(ctx context.Context, fileId string) (*FileShare, error) {
	collection, err := db.GetCollection(ctx, ShareCollectionKey)
	if err != nil {
		return nil, err
	}

	var share FileShare
	err = collection.FindOne(ctx, bson.M{"fileId": fileId}).Decode(&share)

	if err != nil {
		return nil, db.WrapError(errors.WithStack(err), "failed to get share by fileId [%s]", fileId)
	}

	return &share, nil
}

func GetSharedWithUser(ctx context.Context, username string) ([]FileShare, error) {
	collection, err := db.GetCollection(ctx, ShareCollectionKey)
	if err != nil {
		return nil, err
	}

	cursor, err := collection.Find(ctx, bson.M{"accessors": username})
	if err != nil {
		return nil, db.WrapError(err, "failed to get shares for user [%s]", username)
	}

	var shares []FileShare
	err = cursor.All(ctx, &shares)
	if err != nil {
		return nil, db.WrapError(err, "failed to get shares for user [%s]", username)
	}

	return shares, nil

}

func DeleteShare(ctx context.Context, shareId primitive.ObjectID) error {
	collection, err := db.GetCollection(ctx, ShareCollectionKey)
	if err != nil {
		return err
	}

	_, err = collection.DeleteOne(ctx, bson.M{"_id": shareId})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *FileShare) SetPublic(ctx context.Context, pub bool) error {
	collection, err := db.GetCollection(ctx, ShareCollectionKey)
	if err != nil {
		return err
	}

	s.Public = pub

	_, err = collection.UpdateOne(ctx, bson.M{"_id": s.ShareId}, bson.M{"$set": bson.M{"public": pub}})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *FileShare) AddUser(ctx context.Context, username string, perms *Permissions) error {
	collection, err := db.GetCollection(ctx, ShareCollectionKey)
	if err != nil {
		return err
	}

	if s.Permissions == nil {
		s.Permissions = make(map[string]*Permissions)
	}

	// Add new users to the Accessors list and set default permissions
	if !slices.Contains(s.Accessors, username) {
		s.Accessors = append(s.Accessors, username)
	}

	if perms == nil {
		perms = NewPermissions() // default permissions
	}

	if _, ok := s.Permissions[username]; !ok {
		s.Permissions[username] = perms
	}

	// Update the database with the new Accessors list and permissions
	_, err = collection.UpdateOne(ctx, bson.M{"_id": s.ShareId}, bson.M{"$set": bson.M{"accessors": s.Accessors, "permissions": s.Permissions}})
	if err != nil {
		return db.WrapError(err, "failed to add users to share [%s]", s.ShareId)
	}

	return nil
}

func (s *FileShare) SetUserPerms(ctx context.Context, perms map[string]*Permissions) error {
	collection, err := db.GetCollection(ctx, ShareCollectionKey)
	if err != nil {
		return err
	}

	if s.Permissions == nil {
		s.Permissions = make(map[string]*Permissions)
	}

	// Add new users to the Accessors list and set default permissions
	for username, perm := range perms {
		if !slices.Contains(s.Accessors, username) {
			s.Accessors = append(s.Accessors, username)
		}

		if _, ok := s.Permissions[username]; !ok {
			s.Permissions[username] = perm
		}
	}

	// Update the database with the new Accessors list and permissions
	_, err = collection.UpdateOne(ctx, bson.M{"_id": s.ShareId}, bson.M{"$set": bson.M{"accessors": s.Accessors, "permissions": s.Permissions}})
	if err != nil {
		return db.WrapError(err, "failed to add users to share [%s]", s.ShareId)
	}

	return nil
}

func (s *FileShare) RemoveUsers(ctx context.Context, usernames []string) error {
	collection, err := db.GetCollection(ctx, ShareCollectionKey)
	if err != nil {
		return err
	}

	// Remove specified users from the Accessors list
	toRemove := make(map[string]struct{})
	for _, u := range usernames {
		toRemove[u] = struct{}{}
	}

	s.Accessors = slices.Filter(s.Accessors, func(accessor string) bool {
		_, found := toRemove[accessor]
		return !found
	})

	// Remove permissions for these users
	if s.Permissions == nil {
		s.Permissions = make(map[string]*Permissions)
	}

	for _, u := range usernames {
		delete(s.Permissions, u)
	}

	// Update the database with the new Accessors list and permissions
	_, err = collection.UpdateOne(ctx, bson.M{"_id": s.ShareId}, bson.M{"$set": bson.M{"accessors": s.Accessors, "permissions": s.Permissions}})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (s *FileShare) ID() primitive.ObjectID  { return s.ShareId }
func (s *FileShare) GetItemId() string       { return string(s.FileId) }
func (s *FileShare) SetItemId(fileId string) { s.FileId = fileId }
func (s *FileShare) GetAccessors() []string  { return s.Accessors }

// SetUserPermissions sets the permissions for a specific user on this share.
func (s *FileShare) SetUserPermissions(ctx context.Context, username string, perms *Permissions) error {
	if s.Permissions == nil {
		s.Permissions = make(map[string]*Permissions)
	}
	s.Permissions[username] = perms
	collection, err := db.GetCollection(ctx, ShareCollectionKey)
	if err != nil {
		return err
	}
	_, err = collection.UpdateOne(ctx, bson.M{"_id": s.ShareId}, bson.M{"$set": bson.M{"permissions": s.Permissions}})
	return err
}

// GetUserPermissions returns the permissions for a specific user on this share.
func (s *FileShare) GetUserPermissions(username string) *Permissions {
	if s.Permissions == nil {
		return nil
	}
	return s.Permissions[username]
}

// HasPermission checks if a user has a specific permission on this share.
func (s *FileShare) HasPermission(username string, perm Permission) bool {
	if s.Permissions == nil {
		return false
	}
	userPerms := s.Permissions[username]
	if userPerms == nil {
		return false
	}
	return userPerms.HasPermission(perm)
}

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
