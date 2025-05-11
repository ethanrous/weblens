package user

import (
	"context"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/modules/crypto"
	"github.com/ethanrous/weblens/modules/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SaveUser(ctx context.Context, u *User) (err error) {
	if err := validateUsername(u.Username); err != nil {
		return err
	}

	if err := validatePassword(u.Password); err != nil {
		return err
	}

	if u.Password, err = crypto.HashUserPassword(u.Password); err != nil {
		return err
	}

	// if u.HomeId == "" {
	// 	return errors.New("homeId cannot be empty")
	// }

	if u.Id.IsZero() {
		u.Id = primitive.NewObjectID()
	}

	if col, dberr := db.GetCollection(ctx, UserCollectionKey); dberr == nil {
		_, err = col.InsertOne(ctx, u)
	} else {
		return dberr
	}

	return
}

func GetUserByUsername(ctx context.Context, username string) (u *User, err error) {
	col, err := db.GetCollection(ctx, UserCollectionKey)
	if err != nil {
		return
	}

	u = &User{}
	filter := bson.M{"username": username}

	err = col.FindOne(ctx, filter).Decode(u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, errors.WithStack(ErrUserNotFound)
	}

	return
}

func GetAllUsers(ctx context.Context) (us []*User, err error) {
	col, err := db.GetCollection(ctx, UserCollectionKey)
	if err != nil {
		return
	}

	res, err := col.Find(ctx, bson.M{})
	if err != nil {
		return
	}

	err = errors.WithStack(res.All(ctx, &us))

	return
}

func GetServerOwner(ctx context.Context) (u *User, err error) {
	col, err := db.GetCollection(ctx, UserCollectionKey)
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

func (u *User) UpdatePassword(ctx context.Context, newPass string) (err error) {
	if err = validatePassword(newPass); err != nil {
		return
	}

	if u.Password, err = crypto.HashUserPassword(newPass); err != nil {
		return err
	}

	col, err := db.GetCollection(ctx, UserCollectionKey)
	if err != nil {
		return
	}

	_, err = col.UpdateOne(ctx, bson.M{"_id": u.Id}, bson.M{"$set": bson.M{"password": u.Password}})
	if err != nil {
		return errors.WithStack(err)
	}

	return
}

func (u *User) UpdatePermissionLevel(ctx context.Context, newPermissionLevel UserPermissions) (err error) {
	col, err := db.GetCollection(ctx, UserCollectionKey)
	if err != nil {
		return
	}

	_, err = col.UpdateOne(ctx, bson.M{"_id": u.Id}, bson.M{"$set": bson.M{"userPerms": newPermissionLevel}})
	if err != nil {
		return errors.WithStack(err)
	}

	return
}

func (u *User) UpdateHomeId(ctx context.Context, newHomeId string) (err error) {
	col, err := db.GetCollection(ctx, UserCollectionKey)
	if err != nil {
		return
	}

	_, err = col.UpdateOne(ctx, bson.M{"_id": u.Id}, bson.M{"$set": bson.M{"homeId": newHomeId}})
	if err != nil {
		return errors.WithStack(err)
	}

	u.HomeId = newHomeId

	return
}

func (u *User) UpdateTrashId(ctx context.Context, newTrashId string) (err error) {
	col, err := db.GetCollection(ctx, UserCollectionKey)
	if err != nil {
		return
	}

	_, err = col.UpdateOne(ctx, bson.M{"_id": u.Id}, bson.M{"$set": bson.M{"trashId": newTrashId}})
	if err != nil {
		return errors.WithStack(err)
	}

	u.TrashId = newTrashId

	return
}

func (u *User) UpdateActivationStatus(ctx context.Context, active bool) (err error) {
	col, err := db.GetCollection(ctx, UserCollectionKey)
	if err != nil {
		return
	}

	_, err = col.UpdateOne(ctx, bson.M{"_id": u.Id}, bson.M{"$set": bson.M{"activated": active}})
	if err != nil {
		return errors.WithStack(err)
	}

	return
}

func (u *User) UpdateDisplayName(ctx context.Context, newName string) (err error) {
	col, err := db.GetCollection(ctx, UserCollectionKey)
	if err != nil {
		return
	}

	_, err = col.UpdateOne(ctx, bson.M{"_id": u.Id}, bson.M{"$set": bson.M{"fullName": newName}})
	if err != nil {
		return errors.WithStack(err)
	}

	return
}

func (u *User) Delete(ctx context.Context) (err error) {
	col, err := db.GetCollection(ctx, UserCollectionKey)
	if err != nil {
		return
	}

	_, err = col.DeleteOne(ctx, bson.M{"_id": u.Id})
	if err != nil {
		return errors.WithStack(err)
	}

	return
}

func SearchByUsername(ctx context.Context, partialUsername string) ([]*User, error) {
	col, err := db.GetCollection(ctx, UserCollectionKey)
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

func DeleteAllUsers(ctx context.Context) (err error) {
	col, err := db.GetCollection(ctx, UserCollectionKey)
	if err != nil {
		return
	}

	_, err = col.DeleteMany(ctx, bson.M{})
	if err != nil {
		return errors.WithStack(err)
	}

	return
}
