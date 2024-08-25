package database

// func (db *databaseService) GetAllShares() ([]weblens.Share, error) {
// 	ret, err := db.shares.Find(db.ctx, bson.M{})
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var target []*weblens.FileShare
// 	err = ret.All(db.ctx, &target)
//
// 	return internal.SliceConvert[weblens.Share](target), err
// }
//
// func (db *databaseService) UpdateShare(s weblens.Share) error {
// 	filter := bson.M{"_id": s.GetShareId()}
// 	update := bson.M{"$set": s}
// 	o := options.Update().SetUpsert(true)
// 	_, err := db.shares.UpdateOne(db.ctx, filter, update, o)
// 	if err != nil {
// 		return error2.Wrap(err)
// 	}
//
// 	return nil
// }
//
// func (db *databaseService) SetShareEnabledById(sId types.ShareId, enabled bool) error {
// 	return error2.NotImplemented("SetShareEnabledById() has not yet been implemented")
// }
//
// func (db *databaseService) CreateShare(share weblens.Share) error {
// 	_, err := db.shares.InsertOne(db.ctx, share)
// 	return err
// }
//
// func (db *databaseService) AddUsersToShare(share weblens.Share, users []weblens.Username) error {
// 	filter := bson.M{"_id": share.GetShareId()}
// 	update := bson.M{"$addToSet": bson.M{"accessors": bson.M{"$each": users}}}
// 	_, err := db.shares.UpdateOne(db.ctx, filter, update)
// 	return err
// }
//
// func (db *databaseService) GetFileSharesWithUser(username weblens.Username) ([]weblens.Share, error) {
// 	filter := bson.M{"accessors": username, "shareType": "file"}
// 	ret, err := db.shares.Find(db.ctx, filter)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var target []*weblens.FileShare
// 	err = ret.All(db.ctx, &target)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return internal.SliceConvert[weblens.Share](target), nil
// }
//
// func (db *databaseService) DeleteShare(shareId types.ShareId) error {
// 	filter := bson.M{"_id": shareId}
// 	_, err := db.shares.DeleteOne(db.ctx, filter)
// 	return err
// }
