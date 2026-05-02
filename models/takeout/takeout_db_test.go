package takeout_test

import (
	"testing"

	"github.com/ethanrous/weblens/models/db"
	takeout_model "github.com/ethanrous/weblens/models/takeout"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTakeout_SaveAndGet(t *testing.T) {
	ctx := db.SetupTestDB(t, takeout_model.TakeoutCollectionKey, takeout_model.IndexModels...)

	src := []string{"file-id-a", "file-id-b", "file-id-c"}
	require.NoError(t, takeout_model.SaveZip(ctx, takeout_model.NewZip(src, "zip-id-1")))

	got, err := takeout_model.GetZip(ctx, "zip-id-1")
	require.NoError(t, err)
	assert.Equal(t, "zip-id-1", got.TakeoutFileID)
	assert.Equal(t, src, got.SourceFileIDs)
}

func TestTakeout_GetMissingReturnsError(t *testing.T) {
	ctx := db.SetupTestDB(t, takeout_model.TakeoutCollectionKey, takeout_model.IndexModels...)
	_, err := takeout_model.GetZip(ctx, "does-not-exist")
	assert.Error(t, err)
}

func TestTakeout_DuplicateZipFileIDRejected(t *testing.T) {
	ctx := db.SetupTestDB(t, takeout_model.TakeoutCollectionKey, takeout_model.IndexModels...)

	require.NoError(t, takeout_model.SaveZip(ctx, takeout_model.NewZip([]string{"a"}, "dup-zip-id")))
	err := takeout_model.SaveZip(ctx, takeout_model.NewZip([]string{"b", "c"}, "dup-zip-id"))
	assert.Error(t, err, "unique zipFileID index should reject duplicate inserts")
}
