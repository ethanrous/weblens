package service

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestAdjustMediaDates(t *testing.T) {
	type args struct {
		anchor      *models.Media
		newTime     time.Time
		extraMedias []*models.Media
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.wantErr(
					t, AdjustMediaDates(tt.args.anchor, tt.args.newTime, tt.args.extraMedias),
					fmt.Sprintf("AdjustMediaDates(%v, %v, %v)", tt.args.anchor, tt.args.newTime, tt.args.extraMedias),
				)
			},
		)
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

func TestNewShareService(t *testing.T) {
	type args struct {
		collection *mongo.Collection
	}
	tests := []struct {
		name string
		args args
		want models.ShareService
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				assert.Equalf(
					t, tt.want, NewShareService(tt.args.collection), "NewShareService(%v)", tt.args.collection,
				)
			},
		)
	}
}

// func TestNewUserService(t *testing.T) {
// 	type args struct {
// 		col *mongo.Collection
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want *UserServiceImpl
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(
// 			tt.name, func(t *testing.T) {
// 				assert.Equalf(t, tt.want, NewUserService(tt.args.col), "NewUserService(%v)", tt.args.col)
// 			},
// 		)
// 	}
// }

func TestShareServiceImpl_Add(t *testing.T) {
	type fields struct {
		repo   map[models.ShareId]models.Share
		repoMu sync.RWMutex
		col    *mongo.Collection
	}
	type args struct {
		sh models.Share
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ss := &ShareServiceImpl{
					repo:   tt.fields.repo,
					repoMu: tt.fields.repoMu,
					col:    tt.fields.col,
				}
				tt.wantErr(t, ss.Add(tt.args.sh), fmt.Sprintf("Add(%v)", tt.args.sh))
			},
		)
	}
}

func TestShareServiceImpl_AddUsers(t *testing.T) {
	type fields struct {
		repo   map[models.ShareId]models.Share
		repoMu sync.RWMutex
		col    *mongo.Collection
	}
	type args struct {
		share    models.Share
		newUsers []*models.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ss := &ShareServiceImpl{
					repo:   tt.fields.repo,
					repoMu: tt.fields.repoMu,
					col:    tt.fields.col,
				}
				tt.wantErr(
					t, ss.AddUsers(tt.args.share, tt.args.newUsers),
					fmt.Sprintf("AddUsers(%v, %v)", tt.args.share, tt.args.newUsers),
				)
			},
		)
	}
}

func TestShareServiceImpl_Del(t *testing.T) {
	type fields struct {
		repo   map[models.ShareId]models.Share
		repoMu sync.RWMutex
		col    *mongo.Collection
	}
	type args struct {
		sId models.ShareId
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ss := &ShareServiceImpl{
					repo:   tt.fields.repo,
					repoMu: tt.fields.repoMu,
					col:    tt.fields.col,
				}
				tt.wantErr(t, ss.Del(tt.args.sId), fmt.Sprintf("Del(%v)", tt.args.sId))
			},
		)
	}
}

func TestShareServiceImpl_EnableShare(t *testing.T) {
	type fields struct {
		repo   map[models.ShareId]models.Share
		repoMu sync.RWMutex
		col    *mongo.Collection
	}
	type args struct {
		share  models.Share
		enable bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ss := &ShareServiceImpl{
					repo:   tt.fields.repo,
					repoMu: tt.fields.repoMu,
					col:    tt.fields.col,
				}
				tt.wantErr(
					t, ss.EnableShare(tt.args.share, tt.args.enable),
					fmt.Sprintf("EnableShare(%v, %v)", tt.args.share, tt.args.enable),
				)
			},
		)
	}
}

func TestShareServiceImpl_Get(t *testing.T) {
	type fields struct {
		repo   map[models.ShareId]models.Share
		repoMu sync.RWMutex
		col    *mongo.Collection
	}
	type args struct {
		sId models.ShareId
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   models.Share
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ss := &ShareServiceImpl{
					repo:   tt.fields.repo,
					repoMu: tt.fields.repoMu,
					col:    tt.fields.col,
				}
				assert.Equalf(t, tt.want, ss.Get(tt.args.sId), "Get(%v)", tt.args.sId)
			},
		)
	}
}

func TestShareServiceImpl_GetAlbumSharesWithUser(t *testing.T) {
	type fields struct {
		repo   map[models.ShareId]models.Share
		repoMu sync.RWMutex
		col    *mongo.Collection
	}
	type args struct {
		u *models.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*models.AlbumShare
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ss := &ShareServiceImpl{
					repo:   tt.fields.repo,
					repoMu: tt.fields.repoMu,
					col:    tt.fields.col,
				}
				got, err := ss.GetAlbumSharesWithUser(tt.args.u)
				if !tt.wantErr(t, err, fmt.Sprintf("GetAlbumSharesWithUser(%v)", tt.args.u)) {
					return
				}
				assert.Equalf(t, tt.want, got, "GetAlbumSharesWithUser(%v)", tt.args.u)
			},
		)
	}
}

func TestShareServiceImpl_GetAllShares(t *testing.T) {
	type fields struct {
		repo   map[models.ShareId]models.Share
		repoMu sync.RWMutex
		col    *mongo.Collection
	}
	tests := []struct {
		name   string
		fields fields
		want   []models.Share
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ss := &ShareServiceImpl{
					repo:   tt.fields.repo,
					repoMu: tt.fields.repoMu,
					col:    tt.fields.col,
				}
				assert.Equalf(t, tt.want, ss.GetAllShares(), "GetAllShares()")
			},
		)
	}
}

func TestShareServiceImpl_GetFileShare(t *testing.T) {
	type fields struct {
		repo   map[models.ShareId]models.Share
		repoMu sync.RWMutex
		col    *mongo.Collection
	}
	type args struct {
		f *fileTree.WeblensFileImpl
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *models.FileShare
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ss := &ShareServiceImpl{
					repo:   tt.fields.repo,
					repoMu: tt.fields.repoMu,
					col:    tt.fields.col,
				}
				got, err := ss.GetFileShare(tt.args.f)
				if !tt.wantErr(t, err, fmt.Sprintf("GetFileShare(%v)", tt.args.f)) {
					return
				}
				assert.Equalf(t, tt.want, got, "GetFileShare(%v)", tt.args.f)
			},
		)
	}
}

func TestShareServiceImpl_GetFileSharesWithUser(t *testing.T) {
	type fields struct {
		repo   map[models.ShareId]models.Share
		repoMu sync.RWMutex
		col    *mongo.Collection
	}
	type args struct {
		u *models.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*models.FileShare
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ss := &ShareServiceImpl{
					repo:   tt.fields.repo,
					repoMu: tt.fields.repoMu,
					col:    tt.fields.col,
				}
				got, err := ss.GetFileSharesWithUser(tt.args.u)
				if !tt.wantErr(t, err, fmt.Sprintf("GetFileSharesWithUser(%v)", tt.args.u)) {
					return
				}
				assert.Equalf(t, tt.want, got, "GetFileSharesWithUser(%v)", tt.args.u)
			},
		)
	}
}

func TestShareServiceImpl_Init(t *testing.T) {
	type fields struct {
		repo   map[models.ShareId]models.Share
		repoMu sync.RWMutex
		col    *mongo.Collection
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ss := &ShareServiceImpl{
					repo:   tt.fields.repo,
					repoMu: tt.fields.repoMu,
					col:    tt.fields.col,
				}
				tt.wantErr(t, ss.Init(), fmt.Sprintf("Init()"))
			},
		)
	}
}

func TestShareServiceImpl_RemoveUsers(t *testing.T) {
	type fields struct {
		repo   map[models.ShareId]models.Share
		repoMu sync.RWMutex
		col    *mongo.Collection
	}
	type args struct {
		share       models.Share
		removeUsers []*models.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ss := &ShareServiceImpl{
					repo:   tt.fields.repo,
					repoMu: tt.fields.repoMu,
					col:    tt.fields.col,
				}
				tt.wantErr(
					t, ss.RemoveUsers(tt.args.share, tt.args.removeUsers),
					fmt.Sprintf("RemoveUsers(%v, %v)", tt.args.share, tt.args.removeUsers),
				)
			},
		)
	}
}

func TestShareServiceImpl_Size(t *testing.T) {
	type fields struct {
		repo   map[models.ShareId]models.Share
		repoMu sync.RWMutex
		col    *mongo.Collection
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ss := &ShareServiceImpl{
					repo:   tt.fields.repo,
					repoMu: tt.fields.repoMu,
					col:    tt.fields.col,
				}
				assert.Equalf(t, tt.want, ss.Size(), "Size()")
			},
		)
	}
}

func TestShareServiceImpl_writeUpdateTime(t *testing.T) {
	type fields struct {
		repo   map[models.ShareId]models.Share
		repoMu sync.RWMutex
		col    *mongo.Collection
	}
	type args struct {
		sh models.Share
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ss := &ShareServiceImpl{
					repo:   tt.fields.repo,
					repoMu: tt.fields.repoMu,
					col:    tt.fields.col,
				}
				tt.wantErr(t, ss.writeUpdateTime(tt.args.sh), fmt.Sprintf("writeUpdateTime(%v)", tt.args.sh))
			},
		)
	}
}

// func Test_newExif(t *testing.T) {
// 	type args struct {
// 		targetSize  int64
// 		currentSize int64
// 		gexift      *exiftool.Exiftool
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want *exiftool.Exiftool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(
// 			tt.name, func(t *testing.T) {
// 				assert.Equalf(
// 					t, tt.want, newExif(tt.args.targetSize, tt.args.currentSize, tt.args.gexift), "newExif(%v, %v, %v)",
// 					tt.args.targetSize, tt.args.currentSize, tt.args.gexift,
// 				)
// 			},
// 		)
// 	}
// }
