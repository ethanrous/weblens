package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/barasher/go-exiftool"
	"github.com/creativecreature/sturdyc"
	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
	. "github.com/ethrousseau/weblens/service"
	"github.com/ethrousseau/weblens/service/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

type mediaServiceFields struct {
	mediaMap     map[models.ContentId]*models.Media
	streamerMap  map[models.ContentId]*models.VideoStreamer
	exif         *exiftool.Exiftool
	mediaCache   *sturdyc.Client[[]byte]
	typeService  models.MediaTypeService
	fileService  *FileServiceImpl
	albumService *AlbumServiceImpl
	collection   *mongo.Collection
}

type testMedia struct {
	name  string
	media models.Media
	err   error
}

var sampleMediaValid = []testMedia{
	{
		name: "good media",
		media: models.Media{
			ContentId:  "yBjwGUnv5-flkMAmSH-1",
			FileIds:    []fileTree.FileId{"deadbeefdeadbeefdeadbeef"},
			CreateDate: time.Now(),
			Owner:      "weblens",
			Width:      1080,
			Height:     1616,
			PageCount:  1,
			Duration:   0,
			MimeType:   "image/x-sony-arw",
		},
	},
}

var sampleMediaInvalid = []testMedia{
	{
		name: "empty media",
		err:  werror.ErrMediaNoId,
	},
	{
		name: "media missing Id",
		media: models.Media{
			ContentId:  "",
			FileIds:    []fileTree.FileId{"deadbeefdeadbeefdeadbeef"},
			CreateDate: time.Now(),
			Owner:      "weblens",
			Width:      1080,
			Height:     1616,
			PageCount:  1,
			Duration:   0,
			MimeType:   "image/x-sony-arw",
		},
		err: werror.ErrMediaNoId,
	},
	{
		name: "media missing fileIds",
		media: models.Media{
			ContentId:  "yBjwGUnv5-flkMAmSH-3",
			FileIds:    nil,
			CreateDate: time.Now(),
			Owner:      "weblens",
			Width:      1080,
			Height:     1616,
			PageCount:  1,
			Duration:   0,
			MimeType:   "image/x-sony-arw",
		},
		err: werror.ErrMediaNoFiles,
	},
	{
		name: "media missing width",
		media: models.Media{
			ContentId:  "yBjwGUnv5-flkMAmSH-4",
			FileIds:    []fileTree.FileId{"deadbeefdeadbeefdeadbeef"},
			CreateDate: time.Now(),
			Owner:      "weblens",
			Width:      0,
			Height:     1616,
			PageCount:  1,
			Duration:   0,
			MimeType:   "image/x-sony-arw",
		},
		err: werror.ErrMediaNoDimensions,
	},
	{
		name: "image with duration",
		media: models.Media{
			ContentId:  "yBjwGUnv5-flkMAmSH-5",
			FileIds:    []fileTree.FileId{"deadbeefdeadbeefdeadbeef"},
			CreateDate: time.Now(),
			Owner:      "weblens",
			Width:      1080,
			Height:     1616,
			PageCount:  1,
			Duration:   1,
			MimeType:   "image/x-sony-arw",
		},
		err: werror.ErrMediaHasDuration,
	},
	{
		name: "video with no duration",
		media: models.Media{
			ContentId:  "yBjwGUnv5-flkMAmSH-6",
			FileIds:    []fileTree.FileId{"deadbeefdeadbeefdeadbeef"},
			CreateDate: time.Now(),
			Owner:      "weblens",
			Width:      1080,
			Height:     1616,
			PageCount:  1,
			Duration:   0,
			MimeType:   "video/mp4",
		},
		err: werror.ErrMediaNoDuration,
	},
	{
		name: "media bad mime",
		media: models.Media{
			ContentId:  "yBjwGUnv5-flkMAmSH-7",
			FileIds:    []fileTree.FileId{"deadbeefdeadbeefdeadbeef"},
			CreateDate: time.Now(),
			Owner:      "weblens",
			Width:      1080,
			Height:     1616,
			PageCount:  1,
			Duration:   0,
			MimeType:   "itsa me, a mario",
		},
		err: werror.ErrMediaBadMime,
	},
}

var typeService models.MediaTypeService

func TestMediaServiceImpl_Add(t *testing.T) {
	t.Parallel()

	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err != nil {
		panic(err)
	}
	defer col.Drop(context.Background())

	type args struct {
		m *models.Media
	}
	type testArgs struct {
		name    string
		args    args
		wantErr error
	}

	var tests = []testArgs{
		{
			"nil media",
			args{m: nil},
			werror.ErrMediaNil,
		},
	}
	for _, mTest := range sampleMediaValid {
		tests = append(tests, testArgs{mTest.name, args{m: &mTest.media}, mTest.err})
	}
	for _, mTest := range sampleMediaInvalid {
		tests = append(tests, testArgs{mTest.name, args{m: &mTest.media}, mTest.err})
	}

	ms, err := NewMediaService(
		nil, typeService, &mock.MockAlbumService{},
		col,
	)

	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				assert.ErrorIs(t, ms.Add(tt.args.m), tt.wantErr)
			},
		)
	}
}

func TestMediaServiceImpl_Del(t *testing.T) {
	t.Parallel()

	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err != nil {
		panic(err)
	}
	defer col.Drop(context.Background())

	for _, m := range sampleMediaValid {
		_, err := col.InsertOne(context.Background(), &m.media)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
	}

	ms, err := NewMediaService(
		&mock.MockFileService{}, typeService, &mock.MockAlbumService{},
		col,
	)
	require.NoError(t, err)

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

func TestAdjustMediaDates(t *testing.T) {

}
