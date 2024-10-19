package service_test

import (
	"context"
	"testing"

	"github.com/ethanrous/weblens/models"
	. "github.com/ethanrous/weblens/service"
	"github.com/ethanrous/weblens/service/mock"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)

	billUser, err := models.NewUser("billcypher", "shakemyhand", false, true)
	require.NoError(t, err)

	albs := NewAlbumService(col, &mock.MockMediaService{}, ss)

	alb := models.NewAlbum("My precious photos", billUser)

	err = albs.Add(alb)
	require.NoError(t, err)
}
