package service_test

import (
	"context"
	"testing"

	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/service"
	"github.com/ethanrous/weblens/service/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestShareServiceImpl_Add(t *testing.T) {
	t.Parallel()

	logger := log.NewZeroLogger()

	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err != nil {
		panic(err)
	}
	defer col.Drop(context.Background())

	ss, err := service.NewShareService(col, logger)
	if err != nil {
		t.Fatal(err)
	}

	billUser, err := models.NewUser("billcypher", "shakemyhand", "Bill Cypher", false, true)
	require.NoError(t, err)

	dipperUser, err := models.NewUser("dipperpines", "ivegotabook", "Dipper Pines", false, true)
	require.NoError(t, err)

	ft := mock.NewMemFileTree("USERS")
	newDir, err := ft.MkDir(ft.GetRoot(), "billcypher", nil)

	sh := models.NewFileShare(newDir, billUser, []*models.User{dipperUser}, false, false)

	err = ss.Add(sh)
	assert.NoError(t, err)

	// Share does not expand permissions, it has no users, it is not public, etc.
	// So this should not be allowed to be added
	badShare := models.NewFileShare(newDir, billUser, nil, false, false)

	err = ss.Add(badShare)
	assert.Error(t, err)
}

func TestShareServiceImpl_Del(t *testing.T) {
	t.Parallel()

	logger := log.NewZeroLogger()

	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err != nil {
		panic(err)
	}
	defer col.Drop(context.Background())

	ss, err := service.NewShareService(col, logger)
	if err != nil {
		t.Fatal(err)
	}

	billUser, err := models.NewUser("billcypher", "shakemyhand", "Bill Cypher", false, true)
	require.NoError(t, err)

	ft := mock.NewMemFileTree("USERS")
	newDir, err := ft.MkDir(ft.GetRoot(), "billcypher", nil)

	sh := models.NewFileShare(newDir, billUser, nil, true, false)

	err = ss.Add(sh)
	require.NoError(t, err)

	assert.Equal(t, 1, ss.Size())
	docs, err := col.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(1), docs)

	err = ss.Del(sh.ID())
	require.NoError(t, err)

	assert.Equal(t, 0, ss.Size())
	emptyDocs, err := col.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int64(0), emptyDocs)
}

func TestShareServiceImpl_UpdateUsers(t *testing.T) {
	t.Parallel()

	logger := log.NewZeroLogger()

	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err != nil {
		panic(err)
	}
	defer col.Drop(context.Background())

	ss, err := service.NewShareService(col, logger)
	if err != nil {
		t.Fatal(err)
	}

	billUser, err := models.NewUser("billcypher", "shakemyhand", "Bill Cypher", false, true)
	require.NoError(t, err)

	dipperUser, err := models.NewUser("dipperpines", "journalboy123", "Dipper Pines", false, true)
	require.NoError(t, err)

	ft := mock.NewMemFileTree("USERS")
	newDir, err := ft.MkDir(ft.GetRoot(), "billcypher", nil)

	sh := models.NewFileShare(newDir, billUser, nil, true, false)

	err = ss.Add(sh)
	require.NoError(t, err)

	err = ss.AddUsers(sh, []*models.User{dipperUser})
	assert.NoError(t, err)

	assert.Equal(t, 1, len(sh.GetAccessors()))

	err = ss.AddUsers(sh, []*models.User{dipperUser})
	assert.Error(t, err)

	err = ss.RemoveUsers(sh, []*models.User{dipperUser})
	assert.NoError(t, err)

	assert.Equal(t, 0, len(sh.GetAccessors()))

	err = ss.RemoveUsers(sh, []*models.User{dipperUser})
	assert.Error(t, err)
}

// func TestBackupBaseFile(t *testing.T) {
// 	type args struct {
// 		remoteId string
// 		data     []byte
// 		ft       fileTree.FileTree
// 	}
// 	tests := []struct {
// 		name      string
// 		args      args
// 		wantBaseF *fileTree.WeblensFileImpl
// 		wantErr   assert.ErrorAssertionFunc
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(
// 			tt.name, func(t *testing.T) {
// 				gotBaseF, err := BackupBaseFile(tt.args.remoteId, tt.args.data, tt.args.ft)
// 				if !tt.wantErr(
// 					t, err, fmt.Sprintf("BackupBaseFile(%v, %v, %v)", tt.args.remoteId, tt.args.data, tt.args.ft),
// 				) {
// 					return
// 				}
// 				assert.Equalf(
// 					t, tt.wantBaseF, gotBaseF, "BackupBaseFile(%v, %v, %v)", tt.args.remoteId, tt.args.data, tt.args.ft,
// 				)
// 			},
// 		)
// 	}
// }

// func TestNewAccessService(t *testing.T) {
// 	type args struct {
// 		fileService models.FileService
// 		col         *mongo.Collection
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want *AccessServiceImpl
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(
// 			tt.name, func(t *testing.T) {
// 				assert.Equalf(
// 					t, tt.want, NewAccessService(tt.args.fileService, tt.args.col), "NewAccessService(%v, %v)",
// 					tt.args.fileService, tt.args.col,
// 				)
// 			},
// 		)
// 	}
// }

// func TestNewAlbumService(t *testing.T) {
// 	type args struct {
// 		col          *mongo.Collection
// 		mediaService *MediaServiceImpl
// 		shareService models.ShareService
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want *AlbumServiceImpl
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(
// 			tt.name, func(t *testing.T) {
// 				assert.Equalf(
// 					t, tt.want, NewAlbumService(tt.args.col, tt.args.mediaService, tt.args.shareService),
// 					"NewAlbumService(%v, %v, %v)", tt.args.col, tt.args.mediaService, tt.args.shareService,
// 				)
// 			},
// 		)
// 	}
// }

// func TestNewFileService(t *testing.T) {
// 	type args struct {
// 		usersTree     fileTree.FileTree
// 		cacheTree     fileTree.FileTree
// 		userService   models.UserService
// 		accessService models.AccessService
// 		mediaService  models.MediaService
// 		trashCol      *mongo.Collection
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    *FileServiceImpl
// 		wantErr assert.ErrorAssertionFunc
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(
// 			tt.name, func(t *testing.T) {
// 				got, err := NewFileService(
// 					tt.args.usersTree, tt.args.cacheTree, tt.args.userService, tt.args.accessService,
// 					tt.args.mediaService, tt.args.trashCol,
// 				)
// 				if !tt.wantErr(
// 					t, err, fmt.Sprintf(
// 						"NewFileService(%v, %v, %v, %v, %v, %v)", tt.args.usersTree, tt.args.cacheTree,
// 						tt.args.userService, tt.args.accessService, tt.args.mediaService, tt.args.trashCol,
// 					),
// 				) {
// 					return
// 				}
// 				assert.Equalf(
// 					t, tt.want, got, "NewFileService(%v, %v, %v, %v, %v, %v)", tt.args.usersTree, tt.args.cacheTree,
// 					tt.args.userService, tt.args.accessService, tt.args.mediaService, tt.args.trashCol,
// 				)
// 			},
// 		)
// 	}
// }

// func TestNewInstanceService(t *testing.T) {
// 	type args struct {
// 		col *mongo.Collection
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want *InstanceServiceImpl
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(
// 			tt.name, func(t *testing.T) {
// 				assert.Equalf(t, tt.want, NewInstanceService(tt.args.col), "NewInstanceService(%v)", tt.args.col)
// 			},
// 		)
// 	}
// }

// func TestNewMediaService(t *testing.T) {
// 	type args struct {
// 		fileService   *FileServiceImpl
// 		albumService  *AlbumServiceImpl
// 		mediaTypeServ models.MediaTypeService
// 		col           *mongo.Collection
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want *MediaServiceImpl
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(
// 			tt.name, func(t *testing.T) {
// 				assert.Equalf(
// 					t, tt.want,
// 					NewMediaService(tt.args.fileService, tt.args.albumService, tt.args.mediaTypeServ, tt.args.col),
// 					"NewMediaService(%v, %v, %v, %v)", tt.args.fileService, tt.args.albumService, tt.args.mediaTypeServ,
// 					tt.args.col,
// 				)
// 			},
// 		)
// 	}
// }
