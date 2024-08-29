package service

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestAccessServiceImpl_CanUserAccessAlbum(t *testing.T) {
	type fields struct {
		keyMap      map[models.WeblensApiKey]models.ApiKeyInfo
		keyMapMu    *sync.RWMutex
		fileService models.FileService
		collection  *mongo.Collection
	}
	type args struct {
		user  *models.User
		album *models.Album
		share *models.AlbumShare
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				accSrv := &AccessServiceImpl{
					keyMap:      tt.fields.keyMap,
					keyMapMu:    tt.fields.keyMapMu,
					fileService: tt.fields.fileService,
					collection:  tt.fields.collection,
				}
				assert.Equalf(
					t, tt.want, accSrv.CanUserAccessAlbum(tt.args.user, tt.args.album, tt.args.share),
					"CanUserAccessAlbum(%v, %v, %v)", tt.args.user, tt.args.album, tt.args.share,
				)
			},
		)
	}
}

func TestAccessServiceImpl_CanUserAccessFile(t *testing.T) {
	type fields struct {
		keyMap      map[models.WeblensApiKey]models.ApiKeyInfo
		keyMapMu    *sync.RWMutex
		fileService models.FileService
		collection  *mongo.Collection
	}
	type args struct {
		user  *models.User
		file  *fileTree.WeblensFile
		share *models.FileShare
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				accSrv := &AccessServiceImpl{
					keyMap:      tt.fields.keyMap,
					keyMapMu:    tt.fields.keyMapMu,
					fileService: tt.fields.fileService,
					collection:  tt.fields.collection,
				}
				assert.Equalf(
					t, tt.want, accSrv.CanUserAccessFile(tt.args.user, tt.args.file, tt.args.share),
					"CanUserAccessFile(%v, %v, %v)", tt.args.user, tt.args.file, tt.args.share,
				)
			},
		)
	}
}

func TestAccessServiceImpl_CanUserModifyShare(t *testing.T) {
	type fields struct {
		keyMap      map[models.WeblensApiKey]models.ApiKeyInfo
		keyMapMu    *sync.RWMutex
		fileService models.FileService
		collection  *mongo.Collection
	}
	type args struct {
		user  *models.User
		share models.Share
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				accSrv := &AccessServiceImpl{
					keyMap:      tt.fields.keyMap,
					keyMapMu:    tt.fields.keyMapMu,
					fileService: tt.fields.fileService,
					collection:  tt.fields.collection,
				}
				assert.Equalf(
					t, tt.want, accSrv.CanUserModifyShare(tt.args.user, tt.args.share), "CanUserModifyShare(%v, %v)",
					tt.args.user, tt.args.share,
				)
			},
		)
	}
}

func TestAccessServiceImpl_Del(t *testing.T) {
	type fields struct {
		keyMap      map[models.WeblensApiKey]models.ApiKeyInfo
		keyMapMu    *sync.RWMutex
		fileService models.FileService
		collection  *mongo.Collection
	}
	type args struct {
		key models.WeblensApiKey
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
				accSrv := &AccessServiceImpl{
					keyMap:      tt.fields.keyMap,
					keyMapMu:    tt.fields.keyMapMu,
					fileService: tt.fields.fileService,
					collection:  tt.fields.collection,
				}
				tt.wantErr(t, accSrv.Del(tt.args.key), fmt.Sprintf("Del(%v)", tt.args.key))
			},
		)
	}
}

func TestAccessServiceImpl_GenerateApiKey(t *testing.T) {
	type fields struct {
		keyMap      map[models.WeblensApiKey]models.ApiKeyInfo
		keyMapMu    *sync.RWMutex
		fileService models.FileService
		collection  *mongo.Collection
	}
	type args struct {
		creator *models.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    models.ApiKeyInfo
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				accSrv := &AccessServiceImpl{
					keyMap:      tt.fields.keyMap,
					keyMapMu:    tt.fields.keyMapMu,
					fileService: tt.fields.fileService,
					collection:  tt.fields.collection,
				}
				got, err := accSrv.GenerateApiKey(tt.args.creator)
				if !tt.wantErr(t, err, fmt.Sprintf("GenerateApiKey(%v)", tt.args.creator)) {
					return
				}
				assert.Equalf(t, tt.want, got, "GenerateApiKey(%v)", tt.args.creator)
			},
		)
	}
}

func TestAccessServiceImpl_Get(t *testing.T) {
	type fields struct {
		keyMap      map[models.WeblensApiKey]models.ApiKeyInfo
		keyMapMu    *sync.RWMutex
		fileService models.FileService
		collection  *mongo.Collection
	}
	type args struct {
		key models.WeblensApiKey
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    models.ApiKeyInfo
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				accSrv := &AccessServiceImpl{
					keyMap:      tt.fields.keyMap,
					keyMapMu:    tt.fields.keyMapMu,
					fileService: tt.fields.fileService,
					collection:  tt.fields.collection,
				}
				got, err := accSrv.Get(tt.args.key)
				if !tt.wantErr(t, err, fmt.Sprintf("Get(%v)", tt.args.key)) {
					return
				}
				assert.Equalf(t, tt.want, got, "Get(%v)", tt.args.key)
			},
		)
	}
}

func TestAccessServiceImpl_GetAllKeys(t *testing.T) {
	type fields struct {
		keyMap      map[models.WeblensApiKey]models.ApiKeyInfo
		keyMapMu    *sync.RWMutex
		fileService models.FileService
		collection  *mongo.Collection
	}
	type args struct {
		accessor *models.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []models.ApiKeyInfo
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				accSrv := &AccessServiceImpl{
					keyMap:      tt.fields.keyMap,
					keyMapMu:    tt.fields.keyMapMu,
					fileService: tt.fields.fileService,
					collection:  tt.fields.collection,
				}
				got, err := accSrv.GetAllKeys(tt.args.accessor)
				if !tt.wantErr(t, err, fmt.Sprintf("GetAllKeys(%v)", tt.args.accessor)) {
					return
				}
				assert.Equalf(t, tt.want, got, "GetAllKeys(%v)", tt.args.accessor)
			},
		)
	}
}

func TestAccessServiceImpl_Init(t *testing.T) {
	type fields struct {
		keyMap      map[models.WeblensApiKey]models.ApiKeyInfo
		keyMapMu    *sync.RWMutex
		fileService models.FileService
		collection  *mongo.Collection
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
				accSrv := &AccessServiceImpl{
					keyMap:      tt.fields.keyMap,
					keyMapMu:    tt.fields.keyMapMu,
					fileService: tt.fields.fileService,
					collection:  tt.fields.collection,
				}
				tt.wantErr(t, accSrv.Init(), fmt.Sprintf("Init()"))
			},
		)
	}
}

func TestAccessServiceImpl_SetKeyUsedBy(t *testing.T) {
	type fields struct {
		keyMap      map[models.WeblensApiKey]models.ApiKeyInfo
		keyMapMu    *sync.RWMutex
		fileService models.FileService
		collection  *mongo.Collection
	}
	type args struct {
		key    models.WeblensApiKey
		server *models.Instance
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
				accSrv := &AccessServiceImpl{
					keyMap:      tt.fields.keyMap,
					keyMapMu:    tt.fields.keyMapMu,
					fileService: tt.fields.fileService,
					collection:  tt.fields.collection,
				}
				tt.wantErr(
					t, accSrv.SetKeyUsedBy(tt.args.key, tt.args.server),
					fmt.Sprintf("SetKeyUsedBy(%v, %v)", tt.args.key, tt.args.server),
				)
			},
		)
	}
}

func TestAccessServiceImpl_Size(t *testing.T) {
	type fields struct {
		keyMap      map[models.WeblensApiKey]models.ApiKeyInfo
		keyMapMu    *sync.RWMutex
		fileService models.FileService
		collection  *mongo.Collection
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
				accSrv := &AccessServiceImpl{
					keyMap:      tt.fields.keyMap,
					keyMapMu:    tt.fields.keyMapMu,
					fileService: tt.fields.fileService,
					collection:  tt.fields.collection,
				}
				assert.Equalf(t, tt.want, accSrv.Size(), "Size()")
			},
		)
	}
}
