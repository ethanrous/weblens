package database

import (
	"github.com/ethrousseau/weblens/api/dataStore/user"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (db *databaseService) GetAllUsers() ([]types.User, error) {
	ret, err := db.users.Find(db.ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var users []*user.User
	err = ret.All(db.ctx, &users)
	if err != nil {
		return nil, err
	}

	return util.SliceConvert[types.User](users), nil
}

func (db *databaseService) CreateUser(user types.User) error {
	_, err := db.users.InsertOne(db.ctx, user)
	return err
}

func (db *databaseService) UpdatePsaswordByUsername(username types.Username, newPasswordHash string) error {
	return types.NewWeblensError("Not yet implemented")
}

func (db *databaseService) SetAdminByUsername(username types.Username, isAdmin bool) error {
	return types.NewWeblensError("Not yet implemented")
}

func (db *databaseService) ActivateUser(username types.Username) error {
	filter := bson.M{"username": username}
	update := bson.M{"$set": bson.M{"activated": true}}
	_, err := db.users.UpdateOne(db.ctx, filter, update)
	return err
}

func (db *databaseService) AddTokenToUser(username types.Username, token string) error {
	filter := bson.M{"username": username}
	update := bson.M{"$addToSet": bson.M{"tokens": token}}
	_, err := db.users.UpdateOne(db.ctx, filter, update)
	return err
}

func (db *databaseService) SearchUsers(search string) ([]types.Username, error) {
	opts := options.Find().SetProjection(bson.M{"username": 1, "_id": 0})
	ret, err := db.users.Find(db.ctx, bson.M{"username": bson.M{"$regex": search, "$options": "i"}}, opts)

	if err != nil {
		return nil, err
	}

	var users []struct {
		Username string `bson:"username"`
	}
	err = ret.All(db.ctx, &users)
	if err != nil {
		return nil, err
	}

	return util.Map(
		users, func(
			un struct {
				Username string `bson:"username"`
			},
		) types.Username {
			return types.Username(un.Username)
		},
	), nil
}

func (db *databaseService) DeleteAllUsers() error {
	_, err := db.users.DeleteMany(db.ctx, bson.M{})
	return err
}
