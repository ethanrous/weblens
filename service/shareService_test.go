package service_test

import (
	"context"
	"testing"

	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/service"
	"github.com/ethrousseau/weblens/service/mock"
	"github.com/stretchr/testify/assert"
)

func TestShareServiceImplBasic(t *testing.T) {
	t.Parallel()

	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err != nil {
		panic(err)
	}
	defer col.Drop(context.Background())

	ss, err := service.NewShareService(col)
	if err != nil {
		t.Fatal(err)
	}

	ft := mock.NewMemFileTree("MEDIA")
	newDir, err := ft.MkDir(ft.GetRoot(), "billcypher", nil)

	sh := models.NewFileShare(newDir, billUser, []*models.User{dipperUser}, false, false)

	err = ss.Add(sh)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
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
// 		mediaTree     fileTree.FileTree
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
// 					tt.args.mediaTree, tt.args.cacheTree, tt.args.userService, tt.args.accessService,
// 					tt.args.mediaService, tt.args.trashCol,
// 				)
// 				if !tt.wantErr(
// 					t, err, fmt.Sprintf(
// 						"NewFileService(%v, %v, %v, %v, %v, %v)", tt.args.mediaTree, tt.args.cacheTree,
// 						tt.args.userService, tt.args.accessService, tt.args.mediaService, tt.args.trashCol,
// 					),
// 				) {
// 					return
// 				}
// 				assert.Equalf(
// 					t, tt.want, got, "NewFileService(%v, %v, %v, %v, %v, %v)", tt.args.mediaTree, tt.args.cacheTree,
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

