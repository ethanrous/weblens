package services

// import (
// 	"context"
// 	"slices"
// 	"sync"
// 	"time"
//
// 	"github.com/ethanrous/weblens/fileTree"
// 	"github.com/ethanrous/weblens/internal"
// 	"github.com/pkg/errors"
// 	"github.com/ethanrous/weblens/models"
// 	"github.com/rs/zerolog"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/mongo"
// )
//
// type ShareServiceImpl struct {
// 	repo map[models.ShareId]models.Share
//
// 	fileIdMap map[fileTree.FileId]models.ShareId
//
// 	albumIdMap map[models.AlbumId]models.ShareId
//
// 	col    *mongo.Collection
// 	repoMu sync.RWMutex
//
// 	fileMu sync.RWMutex
//
// 	albumMu sync.RWMutex
//
// 	log *zerolog.Logger
// }
//
// func NewShareService(collection *mongo.Collection, log *zerolog.Logger) (models.ShareService, error) {
// 	ss := &ShareServiceImpl{
// 		repo:      make(map[models.ShareId]models.Share),
// 		fileIdMap: make(map[fileTree.FileId]models.ShareId),
// 		col:       collection,
// 	}
//
// 	ret, err := ss.col.Find(context.Background(), bson.M{})
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var target []*models.FileShare
// 	err = ret.All(context.Background(), &target)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	ss.repo = make(map[models.ShareId]models.Share)
// 	for _, sh := range target {
// 		if len(sh.GetAccessors()) == 0 && !sh.IsPublic() && (sh.GetShareType() != models.SharedFile || !sh.IsWormhole()) {
// 			log.Debug().Func(func(e *zerolog.Event) {
// 				e.Msgf("*NOT* Removing %sShare [%s] on init...", sh.GetShareType(), sh.ShareId)
// 			})
// 			// err = db.DeleteShare(sh.GetShareId())
// 			if err != nil {
// 				return nil, err
// 			}
// 			continue
// 		}
//
// 		if sh.Updated.Unix() <= 0 {
// 			sh.UpdatedNow()
// 			err = ss.writeUpdateTime(sh)
// 			if err != nil {
// 				return nil, err
// 			}
// 		}
// 		ss.repo[sh.ID()] = sh
//
// 		if sh.GetShareType() == models.SharedFile {
// 			ss.fileIdMap[sh.FileId] = sh.ID()
// 		}
// 	}
//
// 	return ss, nil
// }
//
// func (ss *ShareServiceImpl) Add(sh models.Share) error {
// 	if len(sh.GetAccessors()) == 0 && !sh.IsPublic() {
// 		return werror.ErrEmptyShare
// 	}
//
// 	_, err := ss.col.InsertOne(context.Background(), sh)
// 	if err != nil {
// 		return err
// 	}
//
// 	ss.repoMu.Lock()
// 	defer ss.repoMu.Unlock()
// 	ss.repo[sh.ID()] = sh
//
// 	if sh.GetShareType() == models.SharedFile {
// 		fileSh, ok := sh.(*models.FileShare)
// 		if !ok {
// 			return werror.ErrBadShareType
// 		}
// 		ss.fileMu.Lock()
// 		defer ss.fileMu.Unlock()
// 		ss.fileIdMap[fileSh.FileId] = sh.ID()
// 	}
//
// 	return nil
// }
//
// func (ss *ShareServiceImpl) Del(sId models.ShareId) error {
// 	if ss.repo[sId] == nil {
// 		return werror.ErrNoShare
// 	}
//
// 	filter := bson.M{"_id": sId}
// 	_, err := ss.col.DeleteOne(context.Background(), filter)
// 	if err != nil {
// 		return err
// 	}
//
// 	ss.repoMu.Lock()
// 	defer ss.repoMu.Unlock()
// 	delete(ss.repo, sId)
// 	return nil
// }
//
// func (ss *ShareServiceImpl) Get(sId models.ShareId) models.Share {
// 	return ss.repo[sId]
// }
//
// func (ss *ShareServiceImpl) GetAllShares() []models.Share {
// 	return internal.MapToValues(ss.repo)
// }
//
// func (ss *ShareServiceImpl) Size() int {
// 	return len(ss.repo)
// }
//
// func (ss *ShareServiceImpl) AddUsers(share models.Share, newUsers []*models.User) error {
// 	addNames := internal.Map(
// 		newUsers, func(u *models.User) string {
// 			return u.GetUsername()
// 		},
// 	)
//
// 	accs := share.GetAccessors()
// 	for _, add := range addNames {
// 		i := slices.Index(accs, add)
// 		if i != -1 {
// 			return werror.ErrUserAlreadyExists
// 		}
// 		accs = append(accs, add)
// 	}
//
// 	filter := bson.M{"_id": share.ID()}
// 	update := bson.M{"$set": bson.M{"accessors": accs, "updated": time.Now()}}
// 	_, err := ss.col.UpdateOne(context.Background(), filter, update)
// 	if err != nil {
// 		return werror.WithStack(err)
// 	}
//
// 	share.SetAccessors(accs)
// 	share.UpdatedNow()
// 	return nil
// }
//
// func (ss *ShareServiceImpl) RemoveUsers(share models.Share, removeUsers []*models.User) error {
// 	removeNames := internal.Map(
// 		removeUsers, func(u *models.User) string {
// 			return u.GetUsername()
// 		},
// 	)
//
// 	accs := share.GetAccessors()
// 	for _, rm := range removeNames {
// 		i := slices.Index(accs, rm)
// 		if i == -1 {
// 			return werror.ErrNoUser
// 		}
//
// 		accs = internal.Banish(accs, i)
// 	}
//
// 	filter := bson.M{"_id": share.ID()}
// 	update := bson.M{"$set": bson.M{"accessors": accs, "updated": time.Now()}}
// 	_, err := ss.col.UpdateOne(context.Background(), filter, update)
// 	if err != nil {
// 		return werror.WithStack(err)
// 	}
//
// 	share.SetAccessors(accs)
// 	share.UpdatedNow()
// 	return nil
// }
//
// func (ss *ShareServiceImpl) GetFileSharesWithUser(u *models.User) ([]*models.FileShare, error) {
// 	filter := bson.M{"accessors": u.GetUsername(), "shareType": "file"}
// 	ret, err := ss.col.Find(context.Background(), filter)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var target []*models.FileShare
// 	err = ret.All(context.Background(), &target)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return target, nil
// }
//
// func (ss *ShareServiceImpl) GetAlbumSharesWithUser(u *models.User) ([]*models.AlbumShare, error) {
// 	filter := bson.M{"accessors": u.GetUsername(), "shareType": "album"}
// 	ret, err := ss.col.Find(context.Background(), filter)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var target []*models.AlbumShare
// 	err = ret.All(context.Background(), &target)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return target, nil
// }
//
// func (ss *ShareServiceImpl) EnableShare(share models.Share, enable bool) error {
// 	_, err := ss.col.UpdateOne(
// 		context.Background(), bson.M{"_id": share.ID()},
// 		bson.M{"$set": bson.M{"enable": enable, "updated": time.Now()}},
// 	)
// 	if err != nil {
// 		return err
// 	}
//
// 	share.SetEnabled(enable)
// 	return nil
// }
//
// func (ss *ShareServiceImpl) SetSharePublic(share models.Share, public bool) error {
// 	_, err := ss.col.UpdateOne(
// 		context.Background(), bson.M{"_id": share.ID()},
// 		bson.M{"$set": bson.M{"public": public, "updated": time.Now()}},
// 	)
// 	if err != nil {
// 		return err
// 	}
//
// 	share.SetPublic(public)
// 	return nil
// }
//
// func (ss *ShareServiceImpl) GetFileShare(fId fileTree.FileId) (*models.FileShare, error) {
// 	ss.fileMu.RLock()
// 	shareId, ok := ss.fileIdMap[fId]
// 	ss.fileMu.RUnlock()
// 	if !ok {
// 		return nil, werror.WithStack(werror.ErrNoShare)
// 	}
//
// 	sh := ss.Get(shareId)
// 	if sh == nil {
// 		return nil, werror.WithStack(werror.ErrExpectedShareMissing)
// 	}
//
// 	fileSh, ok := sh.(*models.FileShare)
// 	if !ok {
// 		return nil, werror.WithStack(werror.ErrBadShareType)
// 	}
//
// 	return fileSh, nil
// }
//
// func (ss *ShareServiceImpl) GetAlbumShare(aId models.AlbumId) (*models.AlbumShare, error) {
// 	ss.albumMu.RLock()
// 	shareId, ok := ss.albumIdMap[aId]
// 	ss.albumMu.RUnlock()
// 	if !ok {
// 		return nil, werror.WithStack(werror.ErrNoShare)
// 	}
//
// 	sh := ss.Get(shareId)
// 	if sh == nil {
// 		return nil, werror.WithStack(werror.ErrExpectedShareMissing)
// 	}
//
// 	albumSh, ok := sh.(*models.AlbumShare)
// 	if !ok {
// 		return nil, werror.WithStack(werror.ErrBadShareType)
// 	}
//
// 	return albumSh, nil
// }
//
// func (ss *ShareServiceImpl) writeUpdateTime(sh models.Share) error {
// 	_, err := ss.col.UpdateOne(
// 		context.Background(), bson.M{"_id": sh.ID()},
// 		bson.M{"$set": bson.M{"updated": sh.LastUpdated()}},
// 	)
//
// 	if err != nil {
// 		return werror.WithStack(err)
// 	}
//
// 	return nil
// }
