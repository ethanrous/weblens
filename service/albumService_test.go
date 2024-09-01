package service_test

import (
	"context"
	"testing"

	"github.com/ethrousseau/weblens/models"
	. "github.com/ethrousseau/weblens/service"
	"github.com/ethrousseau/weblens/service/mock"
	"github.com/stretchr/testify/assert"
)

func TestAlbumServiceImpl_Add(t *testing.T) {
	t.Parallel()

	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err != nil {
		panic(err)
	}
	defer col.Drop(context.Background())

	shareCol := mondb.Collection(t.Name() + "-share")
	err = shareCol.Drop(context.Background())
	if err != nil {
		panic(err)
	}
	defer shareCol.Drop(context.Background())

	ss, err := NewShareService(shareCol)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	albs := NewAlbumService(col, &mock.MockMediaService{}, ss)

	alb := models.NewAlbum("My precious photos", billUser)

	err = albs.Add(alb)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
}
