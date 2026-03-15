package media_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/models/media"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMedia(contentID, owner string) *media.Media {
	m := media.NewMedia(contentID)
	m.SetOwner(owner)
	m.SetEnabled(true)
	m.SetCreateDate(time.Now())

	return m
}

func TestSaveMediaUpsert(t *testing.T) {
	ctx := db.SetupTestDB(t, media.MediaCollectionKey, media.IndexModels...)

	m := newTestMedia("A", "alice")
	require.NoError(t, media.SaveMedia(ctx, m))

	// Upsert with same contentID, different owner — reuse same object to keep _id stable
	m.SetOwner("bob")
	require.NoError(t, media.SaveMedia(ctx, m))

	got, err := media.GetMediaByContentID(ctx, "A")
	require.NoError(t, err)
	assert.Equal(t, "bob", got.GetOwner())

	// Verify only 1 document exists (upsert, not duplicate)
	all, err := media.GetMediasByContentIDs(ctx, "A")
	require.NoError(t, err)
	assert.Len(t, all, 1)
}

func TestGetMediasByContentIDs(t *testing.T) {
	ctx := db.SetupTestDB(t, media.MediaCollectionKey, media.IndexModels...)

	for _, id := range []string{"c1", "c2", "c3"} {
		require.NoError(t, media.SaveMedia(ctx, newTestMedia(id, "alice")))
	}

	t.Run("returns matching subset", func(t *testing.T) {
		got, err := media.GetMediasByContentIDs(ctx, "c1", "c3")
		require.NoError(t, err)
		assert.Len(t, got, 2)
	})

	t.Run("non-existent ID mixed with real", func(t *testing.T) {
		got, err := media.GetMediasByContentIDs(ctx, "c2", "no-such-id")
		require.NoError(t, err)
		assert.Len(t, got, 1)
	})

	t.Run("no IDs returns error", func(t *testing.T) {
		_, err := media.GetMediasByContentIDs(ctx)
		assert.Error(t, err)
	})
}

func TestGetMediasByFileIDs(t *testing.T) {
	ctx := db.SetupTestDB(t, media.MediaCollectionKey, media.IndexModels...)

	m1 := newTestMedia("c1", "alice")
	m1.FileIDs = []string{"f1", "f2"}
	require.NoError(t, media.SaveMedia(ctx, m1))

	m2 := newTestMedia("c2", "alice")
	m2.FileIDs = []string{"f3"}
	require.NoError(t, media.SaveMedia(ctx, m2))

	t.Run("find by f1", func(t *testing.T) {
		got, err := media.GetMediasByFileIDs(ctx, "f1")
		require.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Equal(t, "c1", got[0].ID())
	})

	t.Run("find by f3", func(t *testing.T) {
		got, err := media.GetMediasByFileIDs(ctx, "f3")
		require.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Equal(t, "c2", got[0].ID())
	})

	t.Run("find by f1 and f3", func(t *testing.T) {
		got, err := media.GetMediasByFileIDs(ctx, "f1", "f3")
		require.NoError(t, err)
		assert.Len(t, got, 2)
	})

	t.Run("non-existent fileID", func(t *testing.T) {
		got, err := media.GetMediasByFileIDs(ctx, "no-such-file")
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestGetMediaByOwner(t *testing.T) {
	ctx := db.SetupTestDB(t, media.MediaCollectionKey, media.IndexModels...)

	require.NoError(t, media.SaveMedia(ctx, newTestMedia("a1", "alice")))
	require.NoError(t, media.SaveMedia(ctx, newTestMedia("a2", "alice")))
	require.NoError(t, media.SaveMedia(ctx, newTestMedia("b1", "bob")))

	t.Run("alice has 2", func(t *testing.T) {
		got, err := media.GetMediaByOwner(ctx, "alice")
		require.NoError(t, err)
		assert.Len(t, got, 2)
	})

	t.Run("bob has 1", func(t *testing.T) {
		got, err := media.GetMediaByOwner(ctx, "bob")
		require.NoError(t, err)
		assert.Len(t, got, 1)
	})

	t.Run("charlie has 0", func(t *testing.T) {
		got, err := media.GetMediaByOwner(ctx, "charlie")
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestGetMedia(t *testing.T) {
	ctx := db.SetupTestDB(t, media.MediaCollectionKey, media.IndexModels...)

	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Non-hidden alice media with fileIDs and varying dates
	for i, id := range []string{"m1", "m2", "m3"} {
		m := newTestMedia(id, "alice")
		m.FileIDs = []string{"f-" + id}
		m.SetCreateDate(base.Add(time.Duration(i) * time.Hour))
		require.NoError(t, media.SaveMedia(ctx, m))
	}

	// Hidden alice media
	hidden := newTestMedia("m-hidden", "alice")
	hidden.FileIDs = []string{"f-hidden"}
	hidden.Hidden = true
	hidden.SetCreateDate(base.Add(4 * time.Hour))
	require.NoError(t, media.SaveMedia(ctx, hidden))

	// Alice media with empty fileIDs array (should be excluded by $ne:[])
	noFiles := newTestMedia("m-nofiles", "alice")
	noFiles.FileIDs = []string{}
	noFiles.SetCreateDate(base.Add(5 * time.Hour))
	require.NoError(t, media.SaveMedia(ctx, noFiles))

	// Bob's media
	bobM := newTestMedia("m-bob", "bob")
	bobM.FileIDs = []string{"f-bob"}
	require.NoError(t, media.SaveMedia(ctx, bobM))

	t.Run("non-hidden descending", func(t *testing.T) {
		got, err := media.GetMedia(ctx, "alice", "createDate", -1, nil, false, false)
		require.NoError(t, err)
		assert.Len(t, got, 3)
		// Descending order: m3, m2, m1
		assert.Equal(t, "m3", got[0].ID())
		assert.Equal(t, "m2", got[1].ID())
		assert.Equal(t, "m1", got[2].ID())
	})

	t.Run("allow hidden includes hidden", func(t *testing.T) {
		got, err := media.GetMedia(ctx, "alice", "createDate", -1, nil, false, true)
		require.NoError(t, err)
		assert.Len(t, got, 4)
	})

	t.Run("empty fileIDs excluded", func(t *testing.T) {
		got, err := media.GetMedia(ctx, "alice", "createDate", 1, nil, false, true)
		require.NoError(t, err)

		for _, m := range got {
			assert.NotEqual(t, "m-nofiles", m.ID())
		}
	})
}

func TestGetPagedMedias(t *testing.T) {
	ctx := db.SetupTestDB(t, media.MediaCollectionKey, media.IndexModels...)

	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// 5 image media (duration=0)
	ids := []string{"p1", "p2", "p3", "p4", "p5"}
	for i, id := range ids {
		m := newTestMedia(id, "alice")
		m.MimeType = "image/jpeg"
		m.SetCreateDate(base.Add(time.Duration(i) * time.Hour))
		require.NoError(t, media.SaveMedia(ctx, m))
	}

	// 1 video (duration > 0) — should be excluded
	vid := newTestMedia("vid1", "alice")
	vid.MimeType = "video/mp4"
	vid.Duration = 5000
	vid.SetCreateDate(base.Add(6 * time.Hour))
	require.NoError(t, media.SaveMedia(ctx, vid))

	// 1 raw media
	raw := newTestMedia("raw1", "alice")
	raw.MimeType = "image/x-sony-arw"
	raw.SetCreateDate(base.Add(7 * time.Hour))
	require.NoError(t, media.SaveMedia(ctx, raw))

	allIDs := append(ids, "vid1", "raw1")

	t.Run("page 0 limit 2 ascending", func(t *testing.T) {
		got, err := media.GetPagedMedias(ctx, 2, 0, 1, true, allIDs...)
		require.NoError(t, err)
		assert.Len(t, got, 2)
		assert.Equal(t, "p1", got[0].ID())
		assert.Equal(t, "p2", got[1].ID())
	})

	t.Run("page 1 limit 2 ascending", func(t *testing.T) {
		got, err := media.GetPagedMedias(ctx, 2, 1, 1, true, allIDs...)
		require.NoError(t, err)
		assert.Len(t, got, 2)
		assert.Equal(t, "p3", got[0].ID())
		assert.Equal(t, "p4", got[1].ID())
	})

	t.Run("excludes video", func(t *testing.T) {
		got, err := media.GetPagedMedias(ctx, 100, 0, 1, true, allIDs...)
		require.NoError(t, err)

		for _, m := range got {
			assert.NotEqual(t, "vid1", m.ID())
		}
	})

	t.Run("includeRaw false excludes raw", func(t *testing.T) {
		got, err := media.GetPagedMedias(ctx, 100, 0, 1, false, allIDs...)
		require.NoError(t, err)

		for _, m := range got {
			assert.NotEqual(t, "raw1", m.ID())
		}
	})

	t.Run("includeRaw true includes raw", func(t *testing.T) {
		got, err := media.GetPagedMedias(ctx, 100, 0, 1, true, allIDs...)
		require.NoError(t, err)

		hasRaw := false

		for _, m := range got {
			if m.ID() == "raw1" {
				hasRaw = true
			}
		}

		assert.True(t, hasRaw)
	})
}

func TestGetRandomMedias(t *testing.T) {
	ctx := db.SetupTestDB(t, media.MediaCollectionKey, media.IndexModels...)

	// Insert 10 non-hidden media with fileIDs
	for i := range 10 {
		m := newTestMedia(fmt.Sprintf("rand-%d", i), "alice")
		m.FileIDs = []string{fmt.Sprintf("f-%d", i)}
		m.MimeType = "image/jpeg"
		require.NoError(t, media.SaveMedia(ctx, m))
	}

	t.Run("returns requested count", func(t *testing.T) {
		got, err := media.GetRandomMedias(ctx, media.RandomMediaOptions{Count: 3, Owner: "alice"})
		require.NoError(t, err)
		assert.Len(t, got, 3)
	})

	t.Run("no results for unknown owner", func(t *testing.T) {
		got, err := media.GetRandomMedias(ctx, media.RandomMediaOptions{Count: 3, Owner: "nobody"})
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("NoRaws excludes raw", func(t *testing.T) {
		raw := newTestMedia("raw-rand", "alice")
		raw.FileIDs = []string{"f-raw"}
		raw.MimeType = "image/x-sony-arw"
		require.NoError(t, media.SaveMedia(ctx, raw))

		got, err := media.GetRandomMedias(ctx, media.RandomMediaOptions{Count: 100, Owner: "alice", NoRaws: true})
		require.NoError(t, err)

		for _, m := range got {
			assert.NotEqual(t, "image/x-sony-arw", m.MimeType)
		}
	})
}

func TestDropMediaByOwner(t *testing.T) {
	ctx := db.SetupTestDB(t, media.MediaCollectionKey, media.IndexModels...)

	require.NoError(t, media.SaveMedia(ctx, newTestMedia("a1", "alice")))
	require.NoError(t, media.SaveMedia(ctx, newTestMedia("a2", "alice")))
	require.NoError(t, media.SaveMedia(ctx, newTestMedia("b1", "bob")))

	t.Run("drops only alice media", func(t *testing.T) {
		ids, err := media.DropMediaByOwner(ctx, "alice")
		require.NoError(t, err)
		assert.Len(t, ids, 2)

		// Alice gone
		got, err := media.GetMediaByOwner(ctx, "alice")
		require.NoError(t, err)
		assert.Empty(t, got)

		// Bob still there
		got, err = media.GetMediaByOwner(ctx, "bob")
		require.NoError(t, err)
		assert.Len(t, got, 1)
	})
}

func TestDropMediaByOwnerAll(t *testing.T) {
	ctx := db.SetupTestDB(t, media.MediaCollectionKey, media.IndexModels...)

	require.NoError(t, media.SaveMedia(ctx, newTestMedia("x1", "alice")))
	require.NoError(t, media.SaveMedia(ctx, newTestMedia("x2", "bob")))

	ids, err := media.DropMediaByOwner(ctx, "")
	require.NoError(t, err)
	assert.Nil(t, ids)

	// Everything gone
	got, err := media.GetMediasByContentIDs(ctx, "x1", "x2")
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestDropHDIRs(t *testing.T) {
	ctx := db.SetupTestDB(t, media.MediaCollectionKey, media.IndexModels...)

	m := newTestMedia("hdir-test", "alice")
	m.HDIR = []float64{0.1, 0.2, 0.3}
	require.NoError(t, media.SaveMedia(ctx, m))

	require.NoError(t, media.DropHDIRs(ctx))

	got, err := media.GetMediaByContentID(ctx, "hdir-test")
	require.NoError(t, err)
	assert.Empty(t, got.HDIR)
}

func TestRemoveFileFromMedia(t *testing.T) {
	ctx := db.SetupTestDB(t, media.MediaCollectionKey, media.IndexModels...)

	m := newTestMedia("rf-test", "alice")
	m.FileIDs = []string{"f1", "f2", "f3"}
	require.NoError(t, media.SaveMedia(ctx, m))

	t.Run("removes existing file", func(t *testing.T) {
		require.NoError(t, media.RemoveFileFromMedia(ctx, m, "f2"))

		got, err := media.GetMediaByContentID(ctx, "rf-test")
		require.NoError(t, err)
		assert.Equal(t, []string{"f1", "f3"}, got.GetFiles())
	})

	t.Run("removing non-existent file is no-op", func(t *testing.T) {
		require.NoError(t, media.RemoveFileFromMedia(ctx, m, "no-such"))

		got, err := media.GetMediaByContentID(ctx, "rf-test")
		require.NoError(t, err)
		assert.Equal(t, []string{"f1", "f3"}, got.GetFiles())
	})
}

func TestAddFileToMedia(t *testing.T) {
	ctx := db.SetupTestDB(t, media.MediaCollectionKey, media.IndexModels...)

	m := newTestMedia("af-test", "alice")
	m.FileIDs = []string{"f1"}
	require.NoError(t, media.SaveMedia(ctx, m))

	t.Run("adds new file", func(t *testing.T) {
		require.NoError(t, m.AddFileToMedia(ctx, "f2"))

		got, err := media.GetMediaByContentID(ctx, "af-test")
		require.NoError(t, err)
		assert.Contains(t, got.GetFiles(), "f1")
		assert.Contains(t, got.GetFiles(), "f2")
	})

	t.Run("addToSet prevents duplicates in DB", func(t *testing.T) {
		require.NoError(t, m.AddFileToMedia(ctx, "f1"))

		got, err := media.GetMediaByContentID(ctx, "af-test")
		require.NoError(t, err)
		// DB should have no duplicate thanks to $addToSet
		count := 0

		for _, f := range got.GetFiles() {
			if f == "f1" {
				count++
			}
		}

		assert.Equal(t, 1, count)
	})

	t.Run("in-memory struct also updated", func(t *testing.T) {
		assert.Contains(t, m.GetFiles(), "f2")
	})
}
