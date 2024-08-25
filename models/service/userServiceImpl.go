package service

import (
	"context"
	"iter"
	"maps"
	"sync"

	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type UserServiceImpl struct {
	userMap    map[models.Username]*models.User
	userLock   sync.RWMutex
	publicUser *models.User
	rootUser   *models.User
	col        *mongo.Collection
}

func NewUserService(col *mongo.Collection) *UserServiceImpl {
	return &UserServiceImpl{
		userMap: make(map[models.Username]*models.User),
		col:     col,
	}
}

func (us *UserServiceImpl) Init() error {
	ret, err := us.col.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}
	var users []*models.User
	err = ret.All(context.Background(), &users)
	if err != nil {
		return err
	}

	for _, user := range users {
		us.userMap[user.GetUsername()] = user
	}

	us.publicUser = &models.User{
		Username:   "PUBLIC",
		Activated:  true,
		SystemUser: true,
	}
	us.rootUser = &models.User{
		Username:   "WEBLENS",
		SystemUser: true,
	}

	us.userMap["PUBLIC"] = us.publicUser
	us.userMap["WEBLENS"] = us.rootUser

	return nil
}

func (us *UserServiceImpl) Size() int {
	return len(us.userMap) - 2
}

func (us *UserServiceImpl) GetPublicUser() *models.User {
	return us.publicUser
}

func (us *UserServiceImpl) GetRootUser() *models.User {
	return us.rootUser
}

func (us *UserServiceImpl) Add(user *models.User) error {
	if _, ok := us.userMap[user.GetUsername()]; ok {
		return nil
		// return types.New("user already exists")
	}

	if user.Id == [12]uint8{0} {
		user.Id = primitive.NewObjectID()
	}
	_, err := us.col.InsertOne(context.Background(), user)
	if err != nil {
		return err
	}

	us.userLock.Lock()
	defer us.userLock.Unlock()
	us.userMap[user.GetUsername()] = user

	return nil
}

func (us *UserServiceImpl) Del(un models.Username) error {
	_, err := us.col.DeleteOne(context.Background(), bson.M{"username": un})
	if err != nil {
		return err
	}

	us.userLock.Lock()
	defer us.userLock.Unlock()
	delete(us.userMap, un)

	return nil
}

func (us *UserServiceImpl) ActivateUser(u *models.User) error {
	filter := bson.M{"username": u.GetUsername()}
	update := bson.M{"$set": bson.M{"activated": true}}
	_, err := us.col.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	u.Activated = true

	return nil
}

func (us *UserServiceImpl) GetAll() iter.Seq[*models.User] {
	us.userLock.RLock()
	defer us.userLock.RUnlock()

	return maps.Values(us.userMap)
}

func (us *UserServiceImpl) Get(username models.Username) *models.User {
	us.userLock.RLock()
	defer us.userLock.RUnlock()
	u, ok := us.userMap[username]
	if !ok {
		return nil
	}
	return u
}

func (us *UserServiceImpl) SearchByUsername(searchString string) (iter.Seq[*models.User], error) {
	opts := options.Find().SetProjection(bson.M{"username": 1, "_id": 0}).SetLimit(10)
	ret, err := us.col.Find(
		context.Background(), bson.M{"username": bson.M{"$regex": searchString, "$options": "i"}},
		opts,
	)

	if err != nil {
		return nil, err
	}

	var users []struct {
		Username string `bson:"username"`
	}
	err = ret.All(context.Background(), &users)
	if err != nil {
		return nil, err
	}

	return func(yield func(*models.User) bool) {
		for _, username := range users {
			u := us.Get(models.Username(username.Username))
			if !yield(u) {
				return
			}
		}
	}, nil
}

func (us *UserServiceImpl) SetUserAdmin(u *models.User, admin bool) error {
	filter := bson.M{"username": u.GetUsername()}
	update := bson.M{"$set": bson.M{"admin": admin}}
	_, err := us.col.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	u.Admin = admin

	return nil
}

func (us *UserServiceImpl) UpdateUserPassword(
	username models.Username, oldPassword, newPassword string,
	allowEmptyOld bool,
) error {
	usr := us.userMap[username]

	if !allowEmptyOld || oldPassword != "" {
		if auth := usr.CheckLogin(oldPassword); !auth {
			return werror.ErrBadPassword
		}
	}

	passHashBytes, err := bcrypt.GenerateFromPassword([]byte(newPassword), 11)
	if err != nil {
		return err
	}

	passHashStr := string(passHashBytes)

	filter := bson.M{"username": username}
	update := bson.M{"$set": bson.M{"password": passHashStr}}
	_, err = us.col.UpdateOne(context.Background(), filter, update)

	if err != nil {
		return err
	}

	usr.Password = passHashStr

	return nil
}

func (us *UserServiceImpl) GenerateToken(user *models.User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	tokenString, err := token.SignedString([]byte("key"))
	if err != nil {
		return "", err
	}

	filter := bson.M{"username": user.GetUsername()}
	update := bson.M{"$addToSet": bson.M{"tokens": token}}
	_, err = us.col.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return "", err
	}

	user.AddToken(tokenString)

	return tokenString, nil
}
