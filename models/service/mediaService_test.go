package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/creativecreature/sturdyc"
	"github.com/ethrousseau/weblens/database"
	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/models/mock"
	"github.com/ethrousseau/weblens/models/service"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

type mediaServiceFields struct {
	mediaMap     map[models.ContentId]*models.Media
	streamerMap  map[models.ContentId]*models.VideoStreamer
	exif         *exiftool.Exiftool
	mediaCache   *sturdyc.Client[[]byte]
	typeService  models.MediaTypeService
	fileService  *service.FileServiceImpl
	albumService *service.AlbumServiceImpl
	collection   *mongo.Collection
}

type testMedia struct {
	name  string
	media *models.Media
	err   error
}

var testMimes = map[string]models.MediaType{
	"image/x-sony-arw": {
		Mime:        "image/x-sony-arw",
		Displayable: true,
		Raw:         true,
		Video:       false,
	},
	"video/mp4": {
		Mime:        "video/mp4",
		Displayable: true,
		Raw:         false,
		Video:       true,
	},
}

var mondb *mongo.Database
var sampleMediaValid = []testMedia{
	{
		name: "good media",
		media: &models.Media{
			ContentId:   "yBjwGUnv5-flkMAmSH-1",
			FileIds:     []fileTree.FileId{"deadbeefdeadbeefdeadbeef"},
			CreateDate:  time.Now(),
			Owner:       "weblens",
			MediaWidth:  1080,
			MediaHeight: 1616,
			PageCount:   1,
			Duration:    0,
			MimeType:    "image/x-sony-arw",
		},
	},
}

var sampleMediaInvalid = []testMedia{
	{
		name:  "nil media",
		media: nil,
		err:   werror.ErrMediaNil,
	},
	{
		name: "media missing Id",
		media: &models.Media{
			ContentId:   "",
			FileIds:     []fileTree.FileId{"deadbeefdeadbeefdeadbeef"},
			CreateDate:  time.Now(),
			Owner:       "weblens",
			MediaWidth:  1080,
			MediaHeight: 1616,
			PageCount:   1,
			Duration:    0,
			MimeType:    "image/x-sony-arw",
		},
		err: werror.ErrMediaNoId,
	},
	{
		name: "media missing fileIds",
		media: &models.Media{
			ContentId:   "yBjwGUnv5-flkMAmSH-3",
			FileIds:     nil,
			CreateDate:  time.Now(),
			Owner:       "weblens",
			MediaWidth:  1080,
			MediaHeight: 1616,
			PageCount:   1,
			Duration:    0,
			MimeType:    "image/x-sony-arw",
		},
		err: werror.ErrMediaNoFiles,
	},
	{
		name: "media missing width",
		media: &models.Media{
			ContentId:   "yBjwGUnv5-flkMAmSH-4",
			FileIds:     []fileTree.FileId{"deadbeefdeadbeefdeadbeef"},
			CreateDate:  time.Now(),
			Owner:       "weblens",
			MediaWidth:  0,
			MediaHeight: 1616,
			PageCount:   1,
			Duration:    0,
			MimeType:    "image/x-sony-arw",
		},
		err: werror.ErrMediaNoDimentions,
	},
	{
		name: "image with duration",
		media: &models.Media{
			ContentId:   "yBjwGUnv5-flkMAmSH-5",
			FileIds:     []fileTree.FileId{"deadbeefdeadbeefdeadbeef"},
			CreateDate:  time.Now(),
			Owner:       "weblens",
			MediaWidth:  1080,
			MediaHeight: 1616,
			PageCount:   1,
			Duration:    1,
			MimeType:    "image/x-sony-arw",
		},
		err: werror.ErrMediaHasDuration,
	},
	{
		name: "video with no duration",
		media: &models.Media{
			ContentId:   "yBjwGUnv5-flkMAmSH-6",
			FileIds:     []fileTree.FileId{"deadbeefdeadbeefdeadbeef"},
			CreateDate:  time.Now(),
			Owner:       "weblens",
			MediaWidth:  1080,
			MediaHeight: 1616,
			PageCount:   1,
			Duration:    0,
			MimeType:    "video/mp4",
		},
		err: werror.ErrMediaNoDuration,
	},
	{
		name: "media bad mime",
		media: &models.Media{
			ContentId:   "yBjwGUnv5-flkMAmSH-7",
			FileIds:     []fileTree.FileId{"deadbeefdeadbeefdeadbeef"},
			CreateDate:  time.Now(),
			Owner:       "weblens",
			MediaWidth:  1080,
			MediaHeight: 1616,
			PageCount:   1,
			Duration:    0,
			MimeType:    "itsa me, a mario",
		},
		err: werror.ErrMediaBadMime,
	},
}

func init() {
	if internal.IsDevMode() {
		log.DoDebug()
	}

	mondb = database.ConnectToMongo("mongodb://localhost:27017", "weblens-test")
	if err := mondb.Collection("media").Drop(context.Background()); err != nil {
		panic(err)
	}
}

func TestMediaServiceImpl_Add(t *testing.T) {
	type args struct {
		m *models.Media
	}
	type testArgs struct {
		name    string
		args    args
		wantErr error
	}
	var tests []testArgs
	for _, mTest := range sampleMediaValid {
		tests = append(tests, testArgs{mTest.name, args{m: mTest.media}, mTest.err})
	}
	for _, mTest := range sampleMediaInvalid {
		tests = append(tests, testArgs{mTest.name, args{m: mTest.media}, mTest.err})
	}

	ms, err := service.NewMediaService(
		nil, models.NewTypeService(testMimes),
		mondb.Collection("media"),
	)

	if !assert.NoError(t, err) {
		t.FailNow()
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				assert.ErrorIs(t, ms.Add(tt.args.m), tt.wantErr)
			},
		)
	}

	if err := mondb.Collection("media").Drop(context.Background()); err != nil {
		panic(err)
	}
}

func TestMediaServiceImpl_Del(t *testing.T) {
	for _, m := range sampleMediaValid {
		_, err := mondb.Collection("media").InsertOne(context.Background(), m.media)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
	}

	ms, err := service.NewMediaService(&mock.MockFileService{}, nil, mondb.Collection("media"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.Equal(t, len(sampleMediaValid), ms.Size()) {
		t.FailNow()
	}

	for _, m := range sampleMediaValid {
		err = ms.Del(m.media.ContentId)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
	}

	if !assert.Equal(t, 0, ms.Size()) {
		t.FailNow()
	}
}

func TestMediaServiceImpl_FetchCacheImg(t *testing.T) {
	type args struct {
		m       *models.Media
		q       models.MediaQuality
		pageNum int
	}
	tests := []struct {
		name    string
		fields  mediaServiceFields
		args    args
		want    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ms, _ := service.NewMediaService(nil, nil, nil)
				got, err := ms.FetchCacheImg(tt.args.m, tt.args.q, tt.args.pageNum)
				if !tt.wantErr(
					t, err, fmt.Sprintf("FetchCacheImg(%v, %v, %v)", tt.args.m, tt.args.q, tt.args.pageNum),
				) {
					return
				}
				assert.Equalf(t, tt.want, got, "FetchCacheImg(%v, %v, %v)", tt.args.m, tt.args.q, tt.args.pageNum)
			},
		)
	}
}

func TestMediaServiceImpl_Get(t *testing.T) {
	type args struct {
		mId models.ContentId
	}
	tests := []struct {
		name   string
		fields mediaServiceFields
		args   args
		want   *models.Media
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ms, _ := service.NewMediaService(nil, nil, nil)
				assert.Equalf(t, tt.want, ms.Get(tt.args.mId), "Get(%v)", tt.args.mId)
			},
		)
	}
}

func TestMediaServiceImpl_GetAll(t *testing.T) {

	tests := []struct {
		name   string
		fields mediaServiceFields
		want   []*models.Media
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ms, _ := service.NewMediaService(nil, nil, nil)
				assert.Equalf(t, tt.want, ms.GetAll(), "GetAll()")
			},
		)
	}
}

func TestMediaServiceImpl_GetFilteredMedia(t *testing.T) {

	type args struct {
		requester     *models.User
		sort          string
		sortDirection int
		albumFilter   []models.AlbumId
		allowRaw      bool
		allowHidden   bool
	}
	tests := []struct {
		name    string
		fields  mediaServiceFields
		args    args
		want    []*models.Media
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ms, _ := service.NewMediaService(nil, nil, nil)
				got, err := ms.GetFilteredMedia(
					tt.args.requester, tt.args.sort, tt.args.sortDirection, nil, tt.args.allowRaw,
					tt.args.allowHidden,
				)
				if !tt.wantErr(
					t, err, fmt.Sprintf(
						"GetFilteredMedia(%v, %v, %v, %v, %v, %v)", tt.args.requester, tt.args.sort,
						tt.args.sortDirection, tt.args.albumFilter, tt.args.allowRaw, tt.args.allowHidden,
					),
				) {
					return
				}
				assert.Equalf(
					t, tt.want, got, "GetFilteredMedia(%v, %v, %v, %v, %v, %v)", tt.args.requester, tt.args.sort,
					tt.args.sortDirection, tt.args.albumFilter, tt.args.allowRaw, tt.args.allowHidden,
				)
			},
		)
	}
}

func TestMediaServiceImpl_GetMediaType(t *testing.T) {

	type args struct {
		m *models.Media
	}
	tests := []struct {
		name   string
		fields mediaServiceFields
		args   args
		want   models.MediaType
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ms, _ := service.NewMediaService(nil, nil, nil)
				assert.Equalf(t, tt.want, ms.GetMediaType(tt.args.m), "GetMediaType(%v)", tt.args.m)
			},
		)
	}
}

func TestMediaServiceImpl_GetMediaTypes(t *testing.T) {

	tests := []struct {
		name   string
		fields mediaServiceFields
		want   models.MediaTypeService
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ms, _ := service.NewMediaService(nil, nil, nil)
				assert.Equalf(t, tt.want, ms.GetMediaTypes(), "GetMediaTypes()")
			},
		)
	}
}

func TestMediaServiceImpl_GetProminentColors(t *testing.T) {

	type args struct {
		media *models.Media
	}
	tests := []struct {
		name     string
		fields   mediaServiceFields
		args     args
		wantProm []string
		wantErr  assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ms, _ := service.NewMediaService(nil, nil, nil)
				gotProm, err := ms.GetProminentColors(tt.args.media)
				if !tt.wantErr(t, err, fmt.Sprintf("GetProminentColors(%v)", tt.args.media)) {
					return
				}
				assert.Equalf(t, tt.wantProm, gotProm, "GetProminentColors(%v)", tt.args.media)
			},
		)
	}
}

func TestMediaServiceImpl_HideMedia(t *testing.T) {

	type args struct {
		m      *models.Media
		hidden bool
	}
	tests := []struct {
		name    string
		fields  mediaServiceFields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ms, _ := service.NewMediaService(nil, nil, nil)
				tt.wantErr(
					t, ms.HideMedia(tt.args.m, tt.args.hidden),
					fmt.Sprintf("HideMedia(%v, %v)", tt.args.m, tt.args.hidden),
				)
			},
		)
	}
}

func TestMediaServiceImpl_IsCached(t *testing.T) {

	type args struct {
		m *models.Media
	}
	tests := []struct {
		name   string
		fields mediaServiceFields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ms, _ := service.NewMediaService(nil, nil, nil)
				assert.Equalf(t, tt.want, ms.IsCached(tt.args.m), "IsCached(%v)", tt.args.m)
			},
		)
	}
}

func TestMediaServiceImpl_IsFileDisplayable(t *testing.T) {

	type args struct {
		f *fileTree.WeblensFile
	}
	tests := []struct {
		name   string
		fields mediaServiceFields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ms, _ := service.NewMediaService(nil, nil, nil)
				assert.Equalf(t, tt.want, ms.IsFileDisplayable(tt.args.f), "IsFileDisplayable(%v)", tt.args.f)
			},
		)
	}
}

func TestMediaServiceImpl_LoadMediaFromFile(t *testing.T) {

	type args struct {
		m    *models.Media
		file *fileTree.WeblensFile
	}
	tests := []struct {
		name    string
		fields  mediaServiceFields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ms, _ := service.NewMediaService(nil, nil, nil)
				tt.wantErr(
					t, ms.LoadMediaFromFile(tt.args.m, tt.args.file),
					fmt.Sprintf("LoadMediaFromFile(%v, %v)", tt.args.m, tt.args.file),
				)
			},
		)
	}
}

func TestMediaServiceImpl_NukeCache(t *testing.T) {

	tests := []struct {
		name    string
		fields  mediaServiceFields
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ms, _ := service.NewMediaService(nil, nil, nil)
				tt.wantErr(t, ms.NukeCache(), fmt.Sprintf("NukeCache()"))
			},
		)
	}
}

func TestMediaServiceImpl_RecursiveGetMedia(t *testing.T) {

	type args struct {
		folders []*fileTree.WeblensFile
	}
	tests := []struct {
		name   string
		fields mediaServiceFields
		args   args
		want   []models.ContentId
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// ms, _ := service.NewMediaService( nil, nil, nil)
				// assert.Equalf(
				// 	t, tt.want, ms.RecursiveGetMedia(tt.args.folders...), "RecursiveGetMedia(%v)", tt.args.folders...,
				// )
			},
		)
	}
}

func TestMediaServiceImpl_RemoveFileFromMedia(t *testing.T) {

	type args struct {
		media  *models.Media
		fileId fileTree.FileId
	}
	tests := []struct {
		name    string
		fields  mediaServiceFields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ms, _ := service.NewMediaService(nil, nil, nil)
				tt.wantErr(
					t, ms.RemoveFileFromMedia(tt.args.media, tt.args.fileId),
					fmt.Sprintf("RemoveFileFromMedia(%v, %v)", tt.args.media, tt.args.fileId),
				)
			},
		)
	}
}

func TestMediaServiceImpl_SetMediaLiked(t *testing.T) {

	type args struct {
		mediaId  models.ContentId
		liked    bool
		username models.Username
	}
	tests := []struct {
		name    string
		fields  mediaServiceFields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ms, _ := service.NewMediaService(nil, nil, nil)
				tt.wantErr(
					t, ms.SetMediaLiked(tt.args.mediaId, tt.args.liked, tt.args.username),
					fmt.Sprintf("SetMediaLiked(%v, %v, %v)", tt.args.mediaId, tt.args.liked, tt.args.username),
				)
			},
		)
	}
}

func TestMediaServiceImpl_Size(t *testing.T) {

	tests := []struct {
		name   string
		fields mediaServiceFields
		want   int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ms, _ := service.NewMediaService(nil, nil, nil)
				assert.Equalf(t, tt.want, ms.Size(), "Size()")
			},
		)
	}
}

func TestMediaServiceImpl_StreamCacheVideo(t *testing.T) {

	type args struct {
		m         *models.Media
		startByte int
		endByte   int
	}
	tests := []struct {
		name    string
		fields  mediaServiceFields
		args    args
		want    []byte
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ms, _ := service.NewMediaService(nil, nil, nil)
				got, err := ms.StreamCacheVideo(tt.args.m, tt.args.startByte, tt.args.endByte)
				if !tt.wantErr(
					t, err, fmt.Sprintf("StreamCacheVideo(%v, %v, %v)", tt.args.m, tt.args.startByte, tt.args.endByte),
				) {
					return
				}
				assert.Equalf(
					t, tt.want, got, "StreamCacheVideo(%v, %v, %v)", tt.args.m, tt.args.startByte, tt.args.endByte,
				)
			},
		)
	}
}

func TestMediaServiceImpl_StreamVideo(t *testing.T) {

	type args struct {
		m     *models.Media
		u     *models.User
		share *models.FileShare
	}
	tests := []struct {
		name    string
		fields  mediaServiceFields
		args    args
		want    *models.VideoStreamer
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ms, _ := service.NewMediaService(nil, nil, nil)
				got, err := ms.StreamVideo(tt.args.m, tt.args.u, tt.args.share)
				if !tt.wantErr(t, err, fmt.Sprintf("StreamVideo(%v, %v, %v)", tt.args.m, tt.args.u, tt.args.share)) {
					return
				}
				assert.Equalf(t, tt.want, got, "StreamVideo(%v, %v, %v)", tt.args.m, tt.args.u, tt.args.share)
			},
		)
	}
}

func TestMediaServiceImpl_TypeService(t *testing.T) {

	tests := []struct {
		name   string
		fields mediaServiceFields
		want   models.MediaTypeService
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				ms, _ := service.NewMediaService(nil, nil, nil)
				assert.Equalf(t, tt.want, ms.TypeService(), "TypeService()")
			},
		)
	}
}

// func TestMediaServiceImpl_generateCacheFiles(t *testing.T) {
//
// 	type args struct {
// 		m  *models.Media
// 		bs []byte
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  mediaServiceFields
// 		args    args
// 		want    []*fileTree.WeblensFile
// 		wantErr assert.ErrorAssertionFunc
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(
// 			tt.name, func(t *testing.T) {
// 				ms, _ := service.NewMediaService( nil, nil, nil)
// 				got, err := ms.generateCacheFiles(tt.args.m, tt.args.bs)
// 				if !tt.wantErr(t, err, fmt.Sprintf("generateCacheFiles(%v, %v)", tt.args.m, tt.args.bs)) {
// 					return
// 				}
// 				assert.Equalf(t, tt.want, got, "generateCacheFiles(%v, %v)", tt.args.m, tt.args.bs)
// 			},
// 		)
// 	}
// }

// func TestMediaServiceImpl_getCacheFile(t *testing.T) {
//
// 	type args struct {
// 		m       *models.Media
// 		quality models.MediaQuality
// 		pageNum int
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  mediaServiceFields
// 		args    args
// 		want    *fileTree.WeblensFile
// 		wantErr assert.ErrorAssertionFunc
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(
// 			tt.name, func(t *testing.T) {
// 				ms, _ := service.NewMediaService( nil, nil, nil)
// 				got, err := ms.getCacheFile(tt.args.m, tt.args.quality, tt.args.pageNum)
// 				if !tt.wantErr(
// 					t, err, fmt.Sprintf("getCacheFile(%v, %v, %v)", tt.args.m, tt.args.quality, tt.args.pageNum),
// 				) {
// 					return
// 				}
// 				assert.Equalf(t, tt.want, got, "getCacheFile(%v, %v, %v)", tt.args.m, tt.args.quality, tt.args.pageNum)
// 			},
// 		)
// 	}
// }

// func TestMediaServiceImpl_getFetchMediaCacheImage(t *testing.T) {
//
// 	type args struct {
// 		ctx context.Context
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  mediaServiceFields
// 		args    args
// 		want    []byte
// 		wantErr assert.ErrorAssertionFunc
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(
// 			tt.name, func(t *testing.T) {
// 				ms, _ := service.NewMediaService( nil, nil, nil)
// 				got, err := ms.getFetchMediaCacheImage(tt.args.ctx)
// 				if !tt.wantErr(t, err, fmt.Sprintf("getFetchMediaCacheImage(%v)", tt.args.ctx)) {
// 					return
// 				}
// 				assert.Equalf(t, tt.want, got, "getFetchMediaCacheImage(%v)", tt.args.ctx)
// 			},
// 		)
// 	}
// }

// func TestMediaServiceImpl_removeCacheFiles(t *testing.T) {
//
// 	type args struct {
// 		media *models.Media
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  mediaServiceFields
// 		args    args
// 		wantErr assert.ErrorAssertionFunc
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(
// 			tt.name, func(t *testing.T) {
// 				ms, _ := service.NewMediaService( nil, nil, nil)
// 				tt.wantErr(t, ms.removeCacheFiles(tt.args.media), fmt.Sprintf("removeCacheFiles(%v)", tt.args.media))
// 			},
// 		)
// 	}
// }
