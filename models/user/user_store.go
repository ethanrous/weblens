package user

import (
	"context"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/modules/cryptography"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SaveUser validates and saves a new user to the database, hashing the password before storage.
func SaveUser(ctx context.Context, u *User) (err error) {
	if err := validateUsername(ctx, u.Username); err != nil {
		return err
	}

	if err := validatePassword(u.Password); err != nil {
		return err
	}

	if u.Password, err = cryptography.HashUserPassword(ctx, u.Password); err != nil {
		return err
	}

	// if u.HomeID == "" {
	// 	return errors.New("homeID cannot be empty")
	// }

	if u.ID.IsZero() {
		u.ID = primitive.NewObjectID()
	}

	if col, dberr := db.GetCollection[any](ctx, UserCollectionKey); dberr == nil {
		_, err = col.InsertOne(ctx, u)
	} else {
		return dberr
	}

	return
}

// GetUserByUsername retrieves a user from the database by their username.
func GetUserByUsername(ctx context.Context, username string) (u *User, err error) {
	col, err := db.GetCollection[any](ctx, UserCollectionKey)
	if err != nil {
		return
	}

	u = &User{}
	filter := bson.M{"username": username}

	err = col.FindOne(ctx, filter).Decode(u)
	if err != nil {
		return nil, db.WrapError(err, "failed to get user by username [%s]", username)
	}

	return
}

// DoesUserExist checks if a user with the given username exists in the database.
func DoesUserExist(ctx context.Context, username string) (bool, error) {
	col, err := db.GetCollection[any](ctx, UserCollectionKey)
	if err != nil {
		return false, err
	}

	filter := bson.M{"username": username}

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return false, db.WrapError(err, "failed to check if username [%s] exists", username)
	}

	return count > 0, nil
}

// GetAllUsers retrieves all users from the database.
func GetAllUsers(ctx context.Context) (us []*User, err error) {
	col, err := db.GetCollection[any](ctx, UserCollectionKey)
	if err != nil {
		return
	}

	res, err := col.Find(ctx, bson.M{})
	if err != nil {
		return
	}

	err = wlerrors.WithStack(res.All(ctx, &us))

	return
}

// GetServerOwner retrieves the user with owner permissions from the database.
func GetServerOwner(ctx context.Context) (u *User, err error) {
	col, err := db.GetCollection[any](ctx, UserCollectionKey)
	if err != nil {
		return
	}

	u = &User{}

	// Find all users with Owner permissions
	filter := bson.M{"userPerms": UserPermissionOwner}

	err = col.FindOne(ctx, filter).Decode(u)
	if err != nil {
		return nil, db.WrapError(err, "failed to get server owner")
	}

	return
}

// UpdatePassword validates and updates the user's password, hashing it before storage.
func (u *User) UpdatePassword(ctx context.Context, newPass string) (err error) {
	if err = validatePassword(newPass); err != nil {
		return
	}

	if u.Password, err = cryptography.HashUserPassword(ctx, newPass); err != nil {
		return err
	}

	col, err := db.GetCollection[any](ctx, UserCollectionKey)
	if err != nil {
		return
	}

	_, err = col.UpdateOne(ctx, bson.M{"_id": u.ID}, bson.M{"$set": bson.M{"password": u.Password}})
	if err != nil {
		return wlerrors.WithStack(err)
	}

	return
}

// UpdatePermissionLevel changes the user's permission level in the database.
func (u *User) UpdatePermissionLevel(ctx context.Context, newPermissionLevel Permissions) (err error) {
	col, err := db.GetCollection[any](ctx, UserCollectionKey)
	if err != nil {
		return
	}

	_, err = col.UpdateOne(ctx, bson.M{"_id": u.ID}, bson.M{"$set": bson.M{"userPerms": newPermissionLevel}})
	if err != nil {
		return db.WrapError(err, "failed to update user permission level")
	}

	return
}

// UpdateHomeID updates the user's home directory ID in the database.
func (u *User) UpdateHomeID(ctx context.Context, newHomeID string) (err error) {
	col, err := db.GetCollection[any](ctx, UserCollectionKey)
	if err != nil {
		return
	}

	_, err = col.UpdateOne(ctx, bson.M{"_id": u.ID}, bson.M{"$set": bson.M{"homeID": newHomeID}})
	if err != nil {
		return wlerrors.WithStack(err)
	}

	u.HomeID = newHomeID

	return
}

// UpdateTrashID updates the user's trash directory ID in the database.
func (u *User) UpdateTrashID(ctx context.Context, newTrashID string) (err error) {
	col, err := db.GetCollection[any](ctx, UserCollectionKey)
	if err != nil {
		return
	}

	_, err = col.UpdateOne(ctx, bson.M{"_id": u.ID}, bson.M{"$set": bson.M{"trashID": newTrashID}})
	if err != nil {
		return wlerrors.WithStack(err)
	}

	u.TrashID = newTrashID

	return
}

// UpdateActivationStatus updates the user's account activation status in the database.
func (u *User) UpdateActivationStatus(ctx context.Context, active bool) (err error) {
	col, err := db.GetCollection[any](ctx, UserCollectionKey)
	if err != nil {
		return
	}

	_, err = col.UpdateOne(ctx, bson.M{"_id": u.ID}, bson.M{"$set": bson.M{"activated": active}})
	if err != nil {
		return wlerrors.WithStack(err)
	}

	return
}

// UpdateDisplayName changes the user's display name in the database.
func (u *User) UpdateDisplayName(ctx context.Context, newName string) (err error) {
	col, err := db.GetCollection[any](ctx, UserCollectionKey)
	if err != nil {
		return
	}

	_, err = col.UpdateOne(ctx, bson.M{"_id": u.ID}, bson.M{"$set": bson.M{"fullName": newName}})
	if err != nil {
		return wlerrors.WithStack(err)
	}

	u.DisplayName = newName

	return
}

// Delete removes the user from the database.
func (u *User) Delete(ctx context.Context) (err error) {
	col, err := db.GetCollection[any](ctx, UserCollectionKey)
	if err != nil {
		return
	}

	_, err = col.DeleteOne(ctx, bson.M{"username": u.Username})
	if err != nil {
		return wlerrors.WithStack(err)
	}

	return
}

// SearchByUsername searches for users whose username matches the partial string.
func SearchByUsername(ctx context.Context, partialUsername string) ([]*User, error) {
	col, err := db.GetCollection[any](ctx, UserCollectionKey)
	if err != nil {
		return nil, err
	}

	opts := options.Find().SetLimit(10)

	ret, err := col.Find(context.Background(), bson.M{"username": bson.M{"$regex": partialUsername, "$options": "i"}}, opts)
	if err != nil {
		return nil, err
	}

	var users []*User

	err = ret.All(ctx, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

// DeleteAllUsers removes all users from the database.
func DeleteAllUsers(ctx context.Context) (err error) {
	col, err := db.GetCollection[any](ctx, UserCollectionKey)
	if err != nil {
		return
	}

	_, err = col.DeleteMany(ctx, bson.M{})
	if err != nil {
		return wlerrors.WithStack(err)
	}

	return
}
