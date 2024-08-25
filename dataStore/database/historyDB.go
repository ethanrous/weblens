package database

// func (db *databaseService) WriteFileEvent(fe types.FileEvent) error {
// 	_, err := db.fileHistory.InsertOne(db.ctx, fe)
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
// }
//
// func (db *databaseService) GetAllLifetimes() ([]types.Lifetime, error) {
// 	ret, err := db.fileHistory.Find(db.ctx, bson.M{})
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var target []*fileTree.Lifetime
// 	err = ret.All(db.ctx, &target)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return internal.SliceConvert[types.Lifetime](target), nil
// }
//
// func (db *databaseService) UpsertLifetime(lt types.Lifetime) error {
// 	filter := bson.M{"_id": lt.ID()}
// 	update := bson.M{"$set": lt}
// 	o := options.Update().SetUpsert(true)
// 	_, err := db.fileHistory.UpdateOne(db.ctx, filter, update, o)
//
// 	if err != nil {
// 		return error2.Wrap(err)
// 	}
//
// 	return nil
// }
//
// func (db *databaseService) InsertManyLifetimes(lts []types.Lifetime) error {
// 	_, err := db.fileHistory.InsertMany(db.ctx, internal.SliceConvert[any](lts))
// 	if err != nil {
// 		return error2.Wrap(err)
// 	}
//
// 	return nil
// }
//
// func (db *databaseService) GetActionsByPath(path *fileTree.WeblensFilepath) ([]types.FileAction, error) {
// 	pipe := bson.A{
// 		bson.D{{"$unwind", bson.D{{"path", "$actions"}}}},
// 		bson.D{
// 			{
// 				"$match",
// 				bson.D{
// 					{
// 						"$or",
// 						bson.A{
// 							bson.D{{"actions.originPath", bson.D{{"$regex", path.ToPortable() + "[^/]*/?$"}}}},
// 							bson.D{{"actions.destinationPath", bson.D{{"$regex", path.ToPortable() + "[^/]*/?$"}}}},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		bson.D{{"$replaceRoot", bson.D{{"newRoot", "$actions"}}}},
// 	}
//
// 	ret, err := db.fileHistory.Aggregate(context.TODO(), pipe)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var target []*fileTree.FileAction
// 	err = ret.All(db.ctx, &target)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return internal.SliceConvert[types.FileAction](target), nil
// }
//
// func (db *databaseService) DeleteAllFileHistory() error {
// 	_, err := db.fileHistory.DeleteMany(db.ctx, bson.M{})
// 	return err
// }
//
// func (db *databaseService) GetLifetimesSince(date time.Time) ([]types.Lifetime, error) {
// 	pipe := bson.A{
// 		// bson.D{{"$unwind", bson.D{{"path", "$actions"}}}},
// 		bson.D{
// 			{
// 				"$match",
// 				bson.D{{"actions.timestamp", bson.D{{"$gt", date}}}},
// 			},
// 		},
// 		// bson.D{{"$replaceRoot", bson.D{{"newRoot", "$actions"}}}},
// 		bson.D{{"$sort", bson.D{{"actions.timestamp", 1}}}},
// 	}
// 	ret, err := db.fileHistory.Aggregate(db.ctx, pipe)
// 	if err != nil {
// 		return nil, error2.Wrap(err)
// 	}
//
// 	var target []*fileTree.Lifetime
// 	err = ret.All(db.ctx, &target)
// 	if err != nil {
// 		return nil, error2.Wrap(err)
// 	}
//
// 	return internal.SliceConvert[types.Lifetime](target), nil
// }
//
// func (db *databaseService) GetLatestAction() (types.FileAction, error) {
// 	pipe := bson.A{
// 		bson.D{{"$unwind", bson.D{{"path", "$actions"}}}},
// 		bson.D{{"$sort", bson.D{{"actions.timestamp", -1}}}},
// 		bson.D{{"$limit", 1}},
// 		bson.D{{"$replaceRoot", bson.D{{"newRoot", "$actions"}}}},
// 	}
// 	ret, err := db.fileHistory.Aggregate(db.ctx, pipe)
// 	if err != nil {
// 		return nil, error2.Wrap(err)
// 	}
//
// 	var target []*fileTree.FileAction
// 	err = ret.All(db.ctx, &target)
// 	if err != nil {
// 		return nil, error2.Wrap(err)
// 	}
//
// 	if len(target) == 0 {
// 		return nil, nil
// 	}
//
// 	return target[0], nil
//
// }
