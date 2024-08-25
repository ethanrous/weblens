package database

// func (db *databaseService) GetAllMedia() ([]types.Media, error) {
// 	ret, err := db.media.Find(db.ctx, bson.M{})
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var target []*weblens.Media
// 	err = ret.All(db.ctx, &target)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	medias := internal.SliceConvert[types.Media](target)
//
// 	return medias, nil
// }
//
// func (db *databaseService) CreateMedia(m types.Media) error {
// 	_, err := db.media.InsertOne(db.ctx, m)
// 	return err
// }
//
// func (db *databaseService) AddFileToMedia(mId weblens.ContentId, fId types.FileId) error {
// 	filter := bson.M{"contentId": mId}
// 	update := bson.M{"$addToSet": bson.M{"fileIds": fId}}
// 	_, err := db.media.UpdateOne(db.ctx, filter, update)
// 	return err
// }
//
// func (db *databaseService) RemoveFileFromMedia(mId weblens.ContentId, fId types.FileId) error {
// 	filter := bson.M{"contentId": mId}
// 	update := bson.M{"$pull": bson.M{"fileIds": fId}}
// 	_, err := db.media.UpdateOne(db.ctx, filter, update)
// 	if err != nil {
// 		return error2.Wrap(err)
// 	}
// 	return nil
// }
//
// func (db *databaseService) DeleteMedia(id weblens.ContentId) error {
// 	filter := bson.M{"contentId": id}
// 	_, err := db.media.DeleteOne(db.ctx, filter)
// 	return err
// }
//
// func (db *databaseService) SetMediaHidden(id weblens.ContentId, hidden bool) error {
// 	filter := bson.M{"contentId": id}
// 	_, err := db.media.UpdateOne(db.ctx, filter, bson.M{"$set": bson.M{"hidden": hidden}})
// 	return err
// }
//
// func (db *databaseService) DeleteAllMedia() error {
// 	_, err := db.media.DeleteMany(db.ctx, bson.M{})
// 	return err
// }
//
// func (db *databaseService) GetFetchMediaCacheImage(ctx context.Context) ([]byte, error) {
// 	defer internal.RecoverPanic("Failed to fetch media image into cache")
//
// 	m := ctx.Value("media").(types.Media)
//
// 	q := ctx.Value("quality").(weblens.MediaQuality)
// 	pageNum := ctx.Value("pageNum").(int)
//
// 	f, err := m.GetCacheFile(q, true, pageNum)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	if f == nil {
// 		panic("This should never happen... file is nil in GetFetchMediaCacheImage")
// 	}
//
// 	data, err := f.ReadAll()
// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(data) == 0 {
// 		err = error2.WErrMsg("displayable bytes empty")
// 		return nil, err
// 	}
// 	return data, nil
// }
//
// func (db *databaseService) AddLikeToMedia(id weblens.ContentId, user weblens.Username, liked bool) error {
// 	filter := bson.M{"contentId": id}
// 	var update bson.M
// 	if liked {
// 		update = bson.M{"$addToSet": bson.M{"likedBy": user}}
// 	} else {
// 		update = bson.M{"$pull": bson.M{"likedBy": user}}
// 	}
// 	_, err := db.media.UpdateOne(db.ctx, filter, update)
// 	if err != nil {
// 		return error2.Wrap(err)
// 	}
// 	return nil
// }
