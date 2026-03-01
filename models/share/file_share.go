// Package share provides file sharing functionality and permissions management for the Weblens system.
package share

import (
	"context"
	"time"

	"github.com/ethanrous/weblens/models/db"
	user_model "github.com/ethanrous/weblens/models/usermodel"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/modules/wlog"
	"github.com/ethanrous/weblens/modules/wlslices"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ShareCollectionKey is the MongoDB collection name for shares.
const ShareCollectionKey = "shares"

// ErrShareNotFound is returned when a share cannot be found.
var ErrShareNotFound = wlerrors.New("share not found")

// ErrShareAlreadyExists is returned when attempting to create a share that already exists.
var ErrShareAlreadyExists = wlerrors.New("share already exists")

// FileShare represents a file share configuration.
type FileShare struct {
	// Accessors is a list of users that have access to the share
	Accessors []string `bson:"accessors"`
	// Permissions maps usernames to their specific permissions
	Permissions  map[string]*Permissions `bson:"permissions"`
	Enabled      bool                    `bson:"enabled"`
	Expires      time.Time               `bson:"expires"`
	FileID       string                  `bson:"fileID"`
	Owner        string                  `bson:"owner"`
	Public       bool                    `bson:"public"`
	ShareID      primitive.ObjectID      `bson:"_id"`
	ShareName    string                  `bson:"shareName"`
	Updated      time.Time               `bson:"updated"`
	Wormhole     bool                    `bson:"wormhole"`
	TimelineOnly bool                    `bson:"timelineOnly"`
}

// IndexModels defines MongoDB indexes for the shares collection.
var IndexModels = []mongo.IndexModel{
	{
		Keys: bson.M{
			"fileID": -1,
		},
		Options: options.Index().SetUnique(true),
	},
}

// IDFromString parses a share ID string into a MongoDB ObjectID.
func IDFromString(shareID string) primitive.ObjectID {
	if shareID == "" {
		return primitive.NilObjectID
	}

	id, err := primitive.ObjectIDFromHex(shareID)
	if err != nil {
		wlog.FromContext(context.TODO()).Error().Err(err).Msgf("failed to parse share ID [%s]", shareID)

		return primitive.NilObjectID
	}

	return id
}

// NewFileShare creates a new FileShare with the given parameters.
func NewFileShare(_ context.Context, fileID string, owner *user_model.User, accessors []*user_model.User, public bool, wormhole bool, timelineOnly bool) (*FileShare, error) {
	permissions := make(map[string]*Permissions)
	for _, u := range accessors {
		permissions[u.GetUsername()] = NewPermissions() // default permissions
	}

	return &FileShare{
		FileID: fileID,
		Owner:  owner.GetUsername(),
		Accessors: wlslices.Map(
			accessors, func(u *user_model.User) string {
				return u.GetUsername()
			},
		),
		Permissions:  permissions,
		Public:       public,
		Wormhole:     wormhole,
		TimelineOnly: timelineOnly,
		Enabled:      true,
		Updated:      time.Now(),
	}, nil
}

// SaveFileShare saves a FileShare to the database.
func SaveFileShare(ctx context.Context, share *FileShare) error {
	collection, err := db.GetCollection[any](ctx, ShareCollectionKey)
	if err != nil {
		return err
	}

	if share.ShareID.IsZero() {
		share.ShareID = primitive.NewObjectID()
	}

	if share.Accessors == nil {
		share.Accessors = []string{}
	}

	if share.Permissions == nil {
		share.Permissions = make(map[string]*Permissions)
	}

	_, err = collection.InsertOne(ctx, share)
	if err != nil {
		return db.WrapError(err, "failed to save share [%s]", share.ShareID)
	}

	return nil
}

// GetShareByID retrieves a FileShare by its ID.
func GetShareByID(ctx context.Context, shareID primitive.ObjectID) (*FileShare, error) {
	collection, err := db.GetCollection[any](ctx, ShareCollectionKey)
	if err != nil {
		return nil, err
	}

	share := &FileShare{}

	err = collection.FindOne(ctx, bson.M{"_id": shareID}).Decode(share)
	if err != nil {
		return nil, db.WrapError(wlerrors.WithStack(err), "failed to get share by id [%s]", shareID)
	}

	return share, nil
}

// GetShareByFileID retrieves a FileShare by the file ID it shares.
func GetShareByFileID(ctx context.Context, fileID string) (*FileShare, error) {
	collection, err := db.GetCollection[any](ctx, ShareCollectionKey)
	if err != nil {
		return nil, err
	}

	var share FileShare

	err = collection.FindOne(ctx, bson.M{"fileID": fileID}).Decode(&share)
	if err != nil {
		return nil, db.WrapError(wlerrors.WithStack(err), "failed to get share by fileID [%s]", fileID)
	}

	return &share, nil
}

// GetSharedWithUser retrieves all FileShares that are shared with a specific user.
func GetSharedWithUser(ctx context.Context, username string) ([]FileShare, error) {
	collection, err := db.GetCollection[any](ctx, ShareCollectionKey)
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

// DeleteShare deletes a FileShare from the database.
func DeleteShare(ctx context.Context, shareID primitive.ObjectID) error {
	collection, err := db.GetCollection[any](ctx, ShareCollectionKey)
	if err != nil {
		return err
	}

	_, err = collection.DeleteOne(ctx, bson.M{"_id": shareID})
	if err != nil {
		return wlerrors.WithStack(err)
	}

	return nil
}

// SetPublic sets whether the share is public.
func (s *FileShare) SetPublic(ctx context.Context, pub bool) error {
	collection, err := db.GetCollection[any](ctx, ShareCollectionKey)
	if err != nil {
		return err
	}

	s.Public = pub

	_, err = collection.UpdateOne(ctx, bson.M{"_id": s.ShareID}, bson.M{"$set": bson.M{"public": pub}})
	if err != nil {
		return wlerrors.WithStack(err)
	}

	return nil
}

// AddUser adds a user to the share with the specified permissions.
func (s *FileShare) AddUser(ctx context.Context, username string, perms *Permissions) error {
	collection, err := db.GetCollection[any](ctx, ShareCollectionKey)
	if err != nil {
		return err
	}

	if s.Permissions == nil {
		s.Permissions = make(map[string]*Permissions)
	}

	// Add new users to the Accessors list and set default permissions
	if !wlslices.Contains(s.Accessors, username) {
		s.Accessors = append(s.Accessors, username)
	}

	if perms == nil {
		perms = NewPermissions() // default permissions
	}

	if _, ok := s.Permissions[username]; !ok {
		s.Permissions[username] = perms
	}

	// Update the database with the new Accessors list and permissions
	_, err = collection.UpdateOne(ctx, bson.M{"_id": s.ShareID}, bson.M{"$set": bson.M{"accessors": s.Accessors, "permissions": s.Permissions}})
	if err != nil {
		return db.WrapError(err, "failed to add users to share [%s]", s.ShareID)
	}

	return nil
}

// SetUserPerms sets permissions for multiple users on the share.
func (s *FileShare) SetUserPerms(ctx context.Context, perms map[string]*Permissions) error {
	collection, err := db.GetCollection[any](ctx, ShareCollectionKey)
	if err != nil {
		return err
	}

	if s.Permissions == nil {
		s.Permissions = make(map[string]*Permissions)
	}

	// Add new users to the Accessors list and set default permissions
	for username, perm := range perms {
		if !wlslices.Contains(s.Accessors, username) {
			s.Accessors = append(s.Accessors, username)
		}

		if _, ok := s.Permissions[username]; !ok {
			s.Permissions[username] = perm
		}
	}

	// Update the database with the new Accessors list and permissions
	_, err = collection.UpdateOne(ctx, bson.M{"_id": s.ShareID}, bson.M{"$set": bson.M{"accessors": s.Accessors, "permissions": s.Permissions}})
	if err != nil {
		return db.WrapError(err, "failed to add users to share [%s]", s.ShareID)
	}

	return nil
}

// RemoveUsers removes specified users from the share.
func (s *FileShare) RemoveUsers(ctx context.Context, usernames []string) error {
	collection, err := db.GetCollection[any](ctx, ShareCollectionKey)
	if err != nil {
		return err
	}

	// Remove specified users from the Accessors list
	toRemove := make(map[string]struct{})
	for _, u := range usernames {
		toRemove[u] = struct{}{}
	}

	s.Accessors = wlslices.Filter(s.Accessors, func(accessor string) bool {
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
	_, err = collection.UpdateOne(ctx, bson.M{"_id": s.ShareID}, bson.M{"$set": bson.M{"accessors": s.Accessors, "permissions": s.Permissions}})
	if err != nil {
		return wlerrors.WithStack(err)
	}

	return nil
}

// ID returns the share's unique identifier.
func (s *FileShare) ID() primitive.ObjectID { return s.ShareID }

// GetItemID returns the file ID associated with this share.
func (s *FileShare) GetItemID() string { return string(s.FileID) }

// SetItemID sets the file ID associated with this share.
func (s *FileShare) SetItemID(fileID string) { s.FileID = fileID }

// GetAccessors returns the list of users who have access to this share.
func (s *FileShare) GetAccessors() []string { return s.Accessors }

// SetUserPermissions sets the permissions for a specific user on this share.
func (s *FileShare) SetUserPermissions(ctx context.Context, username string, perms *Permissions) error {
	if s.Permissions == nil {
		s.Permissions = make(map[string]*Permissions)
	}

	s.Permissions[username] = perms

	collection, err := db.GetCollection[any](ctx, ShareCollectionKey)
	if err != nil {
		return err
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": s.ShareID}, bson.M{"$set": bson.M{"permissions": s.Permissions}})

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

// SetAccessors sets the list of users who have access to this share.
func (s *FileShare) SetAccessors(usernames []string) {
	s.Accessors = usernames
}

// GetOwner returns the username of the share's owner.
func (s *FileShare) GetOwner() string { return s.Owner }

// IsPublic returns true if the share is public.
func (s *FileShare) IsPublic() bool { return s.Public }

// IsWormhole returns true if the share is a wormhole share.
func (s *FileShare) IsWormhole() bool { return s.Wormhole }

// IsEnabled returns true if the share is enabled.
func (s *FileShare) IsEnabled() bool { return s.Enabled }

// SetEnabled sets whether the share is enabled.
func (s *FileShare) SetEnabled(enable bool) {
	s.Enabled = enable
}

// UpdatedNow updates the share's last modified timestamp to the current time.
func (s *FileShare) UpdatedNow() {
	s.Updated = time.Now()
}

// LastUpdated returns the last modified timestamp of the share.
func (s *FileShare) LastUpdated() time.Time {
	return s.Updated
}
