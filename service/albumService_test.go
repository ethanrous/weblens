package service

import (
	"fmt"
	"iter"
	"testing"

	"github.com/ethrousseau/weblens/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestAlbumServiceImpl_Add(t *testing.T) {
	type fields struct {
		albumsMap    map[models.AlbumId]*models.Album
		mediaService *MediaServiceImpl
		shareService models.ShareService
		collection   *mongo.Collection
	}
	type args struct {
		a *models.Album
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
				as := &AlbumServiceImpl{
					albumsMap:    tt.fields.albumsMap,
					mediaService: tt.fields.mediaService,
					shareService: tt.fields.shareService,
					collection:   tt.fields.collection,
				}
				tt.wantErr(t, as.Add(tt.args.a), fmt.Sprintf("Add(%v)", tt.args.a))
			},
		)
	}
}

func TestAlbumServiceImpl_AddMediaToAlbum(t *testing.T) {
	type fields struct {
		albumsMap    map[models.AlbumId]*models.Album
		mediaService *MediaServiceImpl
		shareService models.ShareService
		collection   *mongo.Collection
	}
	type args struct {
		album *models.Album
		media []*models.Media
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
				as := &AlbumServiceImpl{
					albumsMap:    tt.fields.albumsMap,
					mediaService: tt.fields.mediaService,
					shareService: tt.fields.shareService,
					collection:   tt.fields.collection,
				}
				tt.wantErr(
					t, as.AddMediaToAlbum(tt.args.album, tt.args.media...),
					fmt.Sprintf("AddMediaToAlbum(%v, %v)", tt.args.album, tt.args.media),
				)
			},
		)
	}
}

func TestAlbumServiceImpl_Del(t *testing.T) {
	type fields struct {
		albumsMap    map[models.AlbumId]*models.Album
		mediaService *MediaServiceImpl
		shareService models.ShareService
		collection   *mongo.Collection
	}
	type args struct {
		aId models.AlbumId
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
				as := &AlbumServiceImpl{
					albumsMap:    tt.fields.albumsMap,
					mediaService: tt.fields.mediaService,
					shareService: tt.fields.shareService,
					collection:   tt.fields.collection,
				}
				tt.wantErr(t, as.Del(tt.args.aId), fmt.Sprintf("Del(%v)", tt.args.aId))
			},
		)
	}
}

func TestAlbumServiceImpl_Get(t *testing.T) {
	type fields struct {
		albumsMap    map[models.AlbumId]*models.Album
		mediaService *MediaServiceImpl
		shareService models.ShareService
		collection   *mongo.Collection
	}
	type args struct {
		aId models.AlbumId
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *models.Album
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				as := &AlbumServiceImpl{
					albumsMap:    tt.fields.albumsMap,
					mediaService: tt.fields.mediaService,
					shareService: tt.fields.shareService,
					collection:   tt.fields.collection,
				}
				assert.Equalf(t, tt.want, as.Get(tt.args.aId), "Get(%v)", tt.args.aId)
			},
		)
	}
}

func TestAlbumServiceImpl_GetAlbumMedias(t *testing.T) {
	type fields struct {
		albumsMap    map[models.AlbumId]*models.Album
		mediaService *MediaServiceImpl
		shareService models.ShareService
		collection   *mongo.Collection
	}
	type args struct {
		album *models.Album
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   iter.Seq[*models.Media]
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				as := &AlbumServiceImpl{
					albumsMap:    tt.fields.albumsMap,
					mediaService: tt.fields.mediaService,
					shareService: tt.fields.shareService,
					collection:   tt.fields.collection,
				}
				assert.Equalf(t, tt.want, as.GetAlbumMedias(tt.args.album), "GetAlbumMedias(%v)", tt.args.album)
			},
		)
	}
}

func TestAlbumServiceImpl_GetAllByUser(t *testing.T) {
	type fields struct {
		albumsMap    map[models.AlbumId]*models.Album
		mediaService *MediaServiceImpl
		shareService models.ShareService
		collection   *mongo.Collection
	}
	type args struct {
		u *models.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*models.Album
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				as := &AlbumServiceImpl{
					albumsMap:    tt.fields.albumsMap,
					mediaService: tt.fields.mediaService,
					shareService: tt.fields.shareService,
					collection:   tt.fields.collection,
				}
				got, err := as.GetAllByUser(tt.args.u)
				if !tt.wantErr(t, err, fmt.Sprintf("GetAllByUser(%v)", tt.args.u)) {
					return
				}
				assert.Equalf(t, tt.want, got, "GetAllByUser(%v)", tt.args.u)
			},
		)
	}
}

func TestAlbumServiceImpl_Init(t *testing.T) {
	type fields struct {
		albumsMap    map[models.AlbumId]*models.Album
		mediaService *MediaServiceImpl
		shareService models.ShareService
		collection   *mongo.Collection
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
				as := &AlbumServiceImpl{
					albumsMap:    tt.fields.albumsMap,
					mediaService: tt.fields.mediaService,
					shareService: tt.fields.shareService,
					collection:   tt.fields.collection,
				}
				tt.wantErr(t, as.Init(), fmt.Sprintf("Init()"))
			},
		)
	}
}

func TestAlbumServiceImpl_RemoveMediaFromAlbum(t *testing.T) {
	type fields struct {
		albumsMap    map[models.AlbumId]*models.Album
		mediaService *MediaServiceImpl
		shareService models.ShareService
		collection   *mongo.Collection
	}
	type args struct {
		album    *models.Album
		mediaIds []models.ContentId
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
				as := &AlbumServiceImpl{
					albumsMap:    tt.fields.albumsMap,
					mediaService: tt.fields.mediaService,
					shareService: tt.fields.shareService,
					collection:   tt.fields.collection,
				}
				tt.wantErr(
					t, as.RemoveMediaFromAlbum(tt.args.album, tt.args.mediaIds...),
					fmt.Sprintf("RemoveMediaFromAlbum(%v, %v)", tt.args.album, tt.args.mediaIds),
				)
			},
		)
	}
}

func TestAlbumServiceImpl_RemoveMediaFromAny(t *testing.T) {
	type fields struct {
		albumsMap    map[models.AlbumId]*models.Album
		mediaService *MediaServiceImpl
		shareService models.ShareService
		collection   *mongo.Collection
	}
	type args struct {
		mediaId models.ContentId
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
				as := &AlbumServiceImpl{
					albumsMap:    tt.fields.albumsMap,
					mediaService: tt.fields.mediaService,
					shareService: tt.fields.shareService,
					collection:   tt.fields.collection,
				}
				tt.wantErr(
					t, as.RemoveMediaFromAny(tt.args.mediaId), fmt.Sprintf("RemoveMediaFromAny(%v)", tt.args.mediaId),
				)
			},
		)
	}
}

func TestAlbumServiceImpl_RenameAlbum(t *testing.T) {
	type fields struct {
		albumsMap    map[models.AlbumId]*models.Album
		mediaService *MediaServiceImpl
		shareService models.ShareService
		collection   *mongo.Collection
	}
	type args struct {
		album   *models.Album
		newName string
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
				as := &AlbumServiceImpl{
					albumsMap:    tt.fields.albumsMap,
					mediaService: tt.fields.mediaService,
					shareService: tt.fields.shareService,
					collection:   tt.fields.collection,
				}
				tt.wantErr(
					t, as.RenameAlbum(tt.args.album, tt.args.newName),
					fmt.Sprintf("RenameAlbum(%v, %v)", tt.args.album, tt.args.newName),
				)
			},
		)
	}
}

func TestAlbumServiceImpl_SetAlbumCover(t *testing.T) {
	type fields struct {
		albumsMap    map[models.AlbumId]*models.Album
		mediaService *MediaServiceImpl
		shareService models.ShareService
		collection   *mongo.Collection
	}
	type args struct {
		albumId models.AlbumId
		cover   *models.Media
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
				as := &AlbumServiceImpl{
					albumsMap:    tt.fields.albumsMap,
					mediaService: tt.fields.mediaService,
					shareService: tt.fields.shareService,
					collection:   tt.fields.collection,
				}
				tt.wantErr(
					t, as.SetAlbumCover(tt.args.albumId, tt.args.cover),
					fmt.Sprintf("SetAlbumCover(%v, %v)", tt.args.albumId, tt.args.cover),
				)
			},
		)
	}
}

func TestAlbumServiceImpl_Size(t *testing.T) {
	type fields struct {
		albumsMap    map[models.AlbumId]*models.Album
		mediaService *MediaServiceImpl
		shareService models.ShareService
		collection   *mongo.Collection
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
				as := &AlbumServiceImpl{
					albumsMap:    tt.fields.albumsMap,
					mediaService: tt.fields.mediaService,
					shareService: tt.fields.shareService,
					collection:   tt.fields.collection,
				}
				assert.Equalf(t, tt.want, as.Size(), "Size()")
			},
		)
	}
}
