package database

// func (db *databaseService) GetAllUsers() ([]*weblens.User, error) {
// 	ret, err := db.users.Find(db.ctx, bson.M{})
// 	if err != nil {
// 		return nil, error2.Wrap(err)
// 	}
// 	var users []*weblens.User
// 	err = ret.All(db.ctx, &users)
// 	if err != nil {
// 		return nil, error2.Wrap(err)
// 	}
//
// 	return internal.SliceConvert[*weblens.User](users), nil
// }
//
// func (db *databaseService) CreateUser(user *weblens.User) error {
// 	_, err := db.users.InsertOne(db.ctx, user)
// 	if err != nil {
// 		return error2.Wrap(err)
// 	}
// 	return nil
// }
//
// func (db *databaseService) UpdatePasswordByUsername(username weblens.Username, newPasswordHash string) error {
// 	filter := bson.M{"username": username}
// 	update := bson.M{"$set": bson.M{"password": newPasswordHash}}
// 	_, err := db.users.UpdateOne(db.ctx, filter, update)
//
// 	if err != nil {
// 		return error2.Wrap(err)
// 	}
//
// 	return nil
// }
//
// func (db *databaseService) SetAdminByUsername(username weblens.Username, isAdmin bool) error {
// 	filter := bson.M{"username": username}
// 	update := bson.M{"$set": bson.M{"admin": isAdmin}}
// 	_, err := db.users.UpdateOne(db.ctx, filter, update)
//
// 	if err != nil {
// 		return error2.Wrap(err)
// 	}
//
// 	return nil
// }
//
// func (db *databaseService) ActivateUser(username weblens.Username) error {
// 	filter := bson.M{"username": username}
// 	update := bson.M{"$set": bson.M{"activated": true}}
// 	_, err := db.users.UpdateOne(db.ctx, filter, update)
// 	return err
// }
//
// func (db *databaseService) AddTokenToUser(username weblens.Username, token string) error {
// 	filter := bson.M{"username": username}
// 	update := bson.M{"$addToSet": bson.M{"tokens": token}}
// 	_, err := db.users.UpdateOne(db.ctx, filter, update)
// 	return err
// }
//
// func (db *databaseService) SearchUsers(search string) ([]weblens.Username, error) {
// 	opts := options.Find().SetProjection(bson.M{"username": 1, "_id": 0})
// 	ret, err := db.users.Find(db.ctx, bson.M{"username": bson.M{"$regex": search, "$options": "i"}}, opts)
//
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var users []struct {
// 		Username string `bson:"username"`
// 	}
// 	err = ret.All(db.ctx, &users)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return internal.Map(
// 		users, func(
// 			un struct {
// 			Username string `bson:"username"`
// 		},
// 		) weblens.Username {
// 			return weblens.Username(un.Username)
// 		},
// 	), nil
// }
//
// func (db *databaseService) DeleteUserByUsername(username weblens.Username) error {
// 	_, err := db.users.DeleteOne(db.ctx, bson.M{"username": username})
// 	return err
// }
//
// func (db *databaseService) DeleteAllUsers() error {
// 	_, err := db.users.DeleteMany(db.ctx, bson.M{})
// 	return err
// }
