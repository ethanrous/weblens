package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ShareServiceImpl struct {
	repo   map[models.ShareId]models.Share
	repoMu sync.RWMutex
	col    *mongo.Collection
}

func NewShareService(collection *mongo.Collection) models.ShareService {
	return &ShareServiceImpl{
		repo: make(map[models.ShareId]models.Share),
		col:  collection,
	}
}

func (ss *ShareServiceImpl) Init() error {
	ret, err := ss.col.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}

	var target []*models.FileShare
	err = ret.All(context.Background(), &target)
	if err != nil {
		return err
	}

	ss.repo = make(map[models.ShareId]models.Share)
	for _, sh := range target {
		if len(sh.GetAccessors()) == 0 && !sh.IsPublic() && (sh.GetShareType() != models.SharedFile || !sh.IsWormhole()) {
			log.Debug.Printf("*NOT* Removing %sShare [%s] on init...", sh.GetShareType(), sh.ShareId)
			// err = db.DeleteShare(sh.GetShareId())
			if err != nil {
				return err
			}
			continue
		}

		if sh.Updated.Unix() <= 0 {
			sh.UpdatedNow()
			ss.writeUpdateTime(sh)
		}
		ss.repo[sh.GetShareId()] = sh
	}

	return nil
}

func (ss *ShareServiceImpl) Add(sh models.Share) error {
	_, err := ss.col.InsertOne(context.Background(), sh)
	if err != nil {
		return err
	}

	ss.repoMu.Lock()
	defer ss.repoMu.Unlock()
	ss.repo[sh.GetShareId()] = sh

	return nil
}

func (ss *ShareServiceImpl) Del(sId models.ShareId) error {
	if ss.repo[sId] == nil {
		return werror.ErrNoShare
	}

	filter := bson.M{"_id": sId}
	_, err := ss.col.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}

	ss.repoMu.Lock()
	defer ss.repoMu.Unlock()
	delete(ss.repo, sId)
	return nil
}

func (ss *ShareServiceImpl) Get(sId models.ShareId) models.Share {
	return ss.repo[sId]
}

func (ss *ShareServiceImpl) GetAllShares() []models.Share {
	return internal.MapToValues(ss.repo)
}

func (ss *ShareServiceImpl) Size() int {
	return len(ss.repo)
}

func (ss *ShareServiceImpl) AddUsers(share models.Share, newUsers []*models.User) error {
	usernames := internal.Map(
		newUsers, func(u *models.User) models.Username {
			return u.GetUsername()
		},
	)

	filter := bson.M{"_id": share.GetShareId()}
	update := bson.M{
		"$addToSet": bson.M{"accessors": bson.M{"$each": usernames}}, "$set": bson.M{"updated": time.Now()},
	}
	_, err := ss.col.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	share.AddUsers(usernames)
	share.UpdatedNow()
	return nil
}

func (ss *ShareServiceImpl) RemoveUsers(share models.Share, removeUsers []*models.User) error {
	usernames := internal.Map(
		removeUsers, func(u *models.User) models.Username {
			return u.GetUsername()
		},
	)

	filter := bson.M{"_id": share.GetShareId()}
	update := bson.M{"$pull": bson.M{"accessors": bson.M{"$each": usernames}}, "$set": bson.M{"updated": time.Now()}}
	_, err := ss.col.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	share.RemoveUsers(usernames)
	return nil
}

func (ss *ShareServiceImpl) GetFileSharesWithUser(u *models.User) ([]*models.FileShare, error) {
	filter := bson.M{"accessors": u.GetUsername(), "shareType": "file"}
	ret, err := ss.col.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	var target []*models.FileShare
	err = ret.All(context.Background(), &target)
	if err != nil {
		return nil, err
	}

	return target, nil
}

func (ss *ShareServiceImpl) GetAlbumSharesWithUser(u *models.User) ([]*models.AlbumShare, error) {
	filter := bson.M{"accessors": u.GetUsername(), "shareType": "album"}
	ret, err := ss.col.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	var target []*models.AlbumShare
	err = ret.All(context.Background(), &target)
	if err != nil {
		return nil, err
	}

	return target, nil
}

func (ss *ShareServiceImpl) EnableShare(share models.Share, enable bool) error {
	_, err := ss.col.UpdateOne(
		context.Background(), bson.M{"_id": share.GetShareId()},
		bson.M{"$set": bson.M{"enable": enable, "updated": time.Now()}},
	)
	if err != nil {
		return err
	}

	share.SetEnabled(enable)
	return nil
}

func (ss *ShareServiceImpl) GetFileShare(f *fileTree.WeblensFile) (*models.FileShare, error) {
	ret := ss.col.FindOne(context.Background(), bson.M{"fileId": f.ID()})
	if ret.Err() != nil {
		if errors.Is(ret.Err(), mongo.ErrNoDocuments) {
			return nil, werror.WithStack(werror.ErrNoShare)
		}
		return nil, werror.WithStack(ret.Err())
	}

	var dbShare models.FileShare
	err := ret.Decode(&dbShare)
	if err != nil {
		return nil, werror.WithStack(err)
	}

	sh := ss.Get(dbShare.ShareId)
	if sh == nil {
		return nil, werror.WithStack(werror.ErrNoShare)
	}

	return sh.(*models.FileShare), nil
}

func (ss *ShareServiceImpl) writeUpdateTime(sh models.Share) error {
	_, err := ss.col.UpdateOne(
		context.Background(), bson.M{"_id": sh.GetShareId()},
		bson.M{"$set": bson.M{"updated": sh.LastUpdated()}},
	)

	if err != nil {
		return werror.WithStack(err)
	}

	return nil
}