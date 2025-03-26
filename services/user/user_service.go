package user

// import (
// 	"context"
// 	"iter"
// 	"sync"
//
// 	"github.com/ethanrous/weblens/database"
// 	"github.com/ethanrous/weblens/internal/werror"
// 	"github.com/ethanrous/weblens/models"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/bson/primitive"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// 	"golang.org/x/crypto/bcrypt"
// )
//
// var _ models.UserService = (*UserServiceImpl)(nil)
//
// type UserServiceImpl struct {
// 	col        database.MongoCollection
// 	userMap    map[string]*models.User
// 	publicUser *models.User
// 	rootUser   *models.User
// 	userLock   sync.RWMutex
// }
//
// func NewUserService(col database.MongoCollection) (*UserServiceImpl, error) {
// 	us := &UserServiceImpl{
// 		userMap: make(map[string]*models.User),
// 		col:     col,
// 	}
//
// 	ret, err := us.col.Find(context.Background(), bson.M{})
// 	if err != nil {
// 		return nil, err
// 	}
// 	var users []*models.User
// 	err = ret.All(context.Background(), &users)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	for _, user := range users {
// 		us.userMap[user.GetUsername()] = user
// 	}
//
// 	us.publicUser = &models.User{
// 		Username:   "PUBLIC",
// 		Activated:  true,
// 		SystemUser: true,
// 	}
// 	us.rootUser = &models.User{
// 		Username:   "WEBLENS",
// 		Admin:      true,
// 		SystemUser: true,
// 	}
//
// 	us.userMap["PUBLIC"] = us.publicUser
// 	us.userMap["WEBLENS"] = us.rootUser
//
// 	return us, nil
// }
//
// func (us *UserServiceImpl) Size() int {
// 	return len(us.userMap) - 2
// }
//
// func (us *UserServiceImpl) GetPublicUser() *models.User {
// 	return us.publicUser
// }
//
// func (us *UserServiceImpl) GetRootUser() *models.User {
// 	return us.rootUser
// }
//
// func (us *UserServiceImpl) Add(user *models.User) error {
// 	if user.GetUsername() == "" {
// 		return werror.Errorf("Cannot add user with no username")
// 	} else if user.PasswordHash == "" {
// 		return werror.Errorf("Cannot add user with no password")
// 	}
//
// 	if _, ok := us.userMap[user.GetUsername()]; ok {
// 		return nil
// 	}
//
// 	if user.Id == [12]uint8{0} {
// 		user.Id = primitive.NewObjectID()
// 	}
//
// 	if user.HomeId == "" || user.TrashId == "" {
// 		return werror.Errorf("Cannot add user [%s] with no home [%s] or trash folder [%s]", user.Username, user.HomeId, user.TrashId)
// 	}
//
// 	_, err := us.col.InsertOne(context.Background(), user)
// 	if err != nil {
// 		return err
// 	}
//
// 	us.userLock.Lock()
// 	defer us.userLock.Unlock()
// 	us.userMap[user.GetUsername()] = user
//
// 	return nil
// }
//
// func (us *UserServiceImpl) CreateOwner(username, password, fullname string) (*models.User, error) {
// 	owner, err := models.NewUser(username, password, fullname, true, true)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	owner.IsServerOwner = true
//
// 	_, err = us.col.InsertOne(context.Background(), owner)
// 	if err != nil {
// 		return nil, werror.WithStack(err)
// 	}
//
// 	us.userMap[username] = owner
//
// 	return owner, nil
// }
//
// func (us *UserServiceImpl) Del(un string) error {
// 	_, err := us.col.DeleteOne(context.Background(), bson.M{"username": un})
// 	if err != nil {
// 		return err
// 	}
//
// 	us.userLock.Lock()
// 	defer us.userLock.Unlock()
// 	delete(us.userMap, un)
//
// 	return nil
// }
//
// func (us *UserServiceImpl) ActivateUser(u *models.User, active bool) error {
// 	filter := bson.M{"username": u.GetUsername()}
// 	update := bson.M{"$set": bson.M{"activated": active}}
// 	_, err := us.col.UpdateOne(context.Background(), filter, update)
// 	if err != nil {
// 		return err
// 	}
//
// 	u.Activated = active
//
// 	return nil
// }
//
// func (us *UserServiceImpl) UpdateFullName(u *models.User, newFullName string) error {
// 	filter := bson.M{"username": u.GetUsername()}
// 	update := bson.M{"$set": bson.M{"fullName": newFullName}}
// 	_, err := us.col.UpdateOne(context.Background(), filter, update)
// 	if err != nil {
// 		return werror.WithStack(err)
// 	}
//
// 	u.SetFullName(newFullName)
//
// 	return nil
// }
//
// func (us *UserServiceImpl) GetAll() (iter.Seq[*models.User], error) {
// 	us.userLock.RLock()
// 	defer us.userLock.RUnlock()
//
// 	return func(yield func(user *models.User) bool) {
// 		for _, u := range us.userMap {
// 			if u.SystemUser {
// 				continue
// 			}
// 			if !yield(u) {
// 				return
// 			}
// 		}
// 	}, nil
// }
//
// func (us *UserServiceImpl) Get(username string) *models.User {
// 	us.userLock.RLock()
// 	defer us.userLock.RUnlock()
// 	u, ok := us.userMap[username]
// 	if !ok {
// 		return nil
// 	}
// 	return u
// }
//
// func (us *UserServiceImpl) SearchByUsername(searchString string) (iter.Seq[*models.User], error) {
// 	opts := options.Find().SetProjection(bson.M{"username": 1, "_id": 0}).SetLimit(10)
// 	ret, err := us.col.Find(
// 		context.Background(), bson.M{"username": bson.M{"$regex": searchString, "$options": "i"}},
// 		opts,
// 	)
//
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	var users []struct {
// 		Username string `bson:"username"`
// 	}
// 	err = ret.All(context.Background(), &users)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return func(yield func(*models.User) bool) {
// 		for _, username := range users {
// 			u := us.Get(username.Username)
// 			if !yield(u) {
// 				return
// 			}
// 		}
// 	}, nil
// }
//
// func (us *UserServiceImpl) SetUserAdmin(u *models.User, admin bool) error {
// 	if !u.IsActive() {
// 		return werror.WithStack(werror.ErrUserNotActive)
// 	}
//
// 	filter := bson.M{"username": u.GetUsername()}
// 	update := bson.M{"$set": bson.M{"admin": admin}}
// 	_, err := us.col.UpdateOne(context.Background(), filter, update)
// 	if err != nil {
// 		return err
// 	}
//
// 	u.Admin = admin
//
// 	return nil
// }
//
// func (us *UserServiceImpl) UpdateUserPassword(
// 	username string, oldPassword, newPassword string,
// 	allowEmptyOld bool,
// ) error {
// 	usr := us.userMap[username]
//
// 	if !allowEmptyOld || oldPassword != "" {
// 		if auth := usr.CheckLogin(oldPassword); !auth {
// 			return werror.ErrBadPassword
// 		}
// 	}
//
// 	passHashBytes, err := bcrypt.GenerateFromPassword([]byte(newPassword), 11)
// 	if err != nil {
// 		return err
// 	}
//
// 	passHashStr := string(passHashBytes)
//
// 	filter := bson.M{"username": username}
// 	update := bson.M{"$set": bson.M{"password": passHashStr}}
// 	_, err = us.col.UpdateOne(context.Background(), filter, update)
//
// 	if err != nil {
// 		return err
// 	}
//
// 	usr.PasswordHash = passHashStr
//
// 	return nil
// }
//
// func (us *UserServiceImpl) UpdateUserHome(u *models.User) error {
// 	_, err := us.col.UpdateOne(
// 		context.Background(), bson.M{"username": u.GetUsername()},
// 		bson.M{"$set": bson.M{"homeId": u.HomeId, "trashId": u.TrashId}},
// 	)
// 	if err != nil {
// 		return werror.WithStack(err)
// 	}
//
// 	return nil
// }
