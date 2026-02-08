package media_test

import (
	"testing"
	"time"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/models/media"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMedia(t *testing.T) {
	t.Run("creates media with content ID", func(t *testing.T) {
		contentID := "test-content-id"
		m := media.NewMedia(contentID)
		assert.NotNil(t, m)
		assert.Equal(t, contentID, m.ID())
	})
}

func TestMediaID(t *testing.T) {
	t.Run("returns content ID", func(t *testing.T) {
		m := media.NewMedia("my-content-id")
		assert.Equal(t, "my-content-id", m.ID())
	})
}

func TestMediaSetContentID(t *testing.T) {
	t.Run("sets content ID", func(t *testing.T) {
		m := media.NewMedia("original")
		m.SetContentID("updated")
		assert.Equal(t, "updated", m.ID())
	})
}

func TestMediaIsHidden(t *testing.T) {
	t.Run("returns hidden status", func(t *testing.T) {
		m := media.NewMedia("test")
		assert.False(t, m.IsHidden())
	})
}

func TestMediaCreateDate(t *testing.T) {
	t.Run("sets and gets create date", func(t *testing.T) {
		m := media.NewMedia("test")
		now := time.Now()
		m.SetCreateDate(now)
		assert.Equal(t, now.Unix(), m.GetCreateDate().Unix())
	})
}

func TestMediaPageCount(t *testing.T) {
	t.Run("returns page count", func(t *testing.T) {
		m := media.NewMedia("test")
		assert.Equal(t, 0, m.GetPageCount())
	})
}

func TestMediaVideoLength(t *testing.T) {
	t.Run("returns video length", func(t *testing.T) {
		m := media.NewMedia("test")
		assert.Equal(t, 0, m.GetVideoLength())
	})
}

func TestMediaOwner(t *testing.T) {
	t.Run("sets and gets owner", func(t *testing.T) {
		m := media.NewMedia("test")
		m.SetOwner("testuser")
		assert.Equal(t, "testuser", m.GetOwner())
	})
}

func TestMediaFiles(t *testing.T) {
	t.Run("returns empty file list initially", func(t *testing.T) {
		m := media.NewMedia("test")
		assert.Empty(t, m.GetFiles())
	})
}

func TestMediaImported(t *testing.T) {
	t.Run("sets and gets imported status", func(t *testing.T) {
		m := media.NewMedia("test")
		assert.False(t, m.IsImported())
		m.SetImported(true)
		assert.True(t, m.IsImported())
	})

	t.Run("returns false for nil media", func(t *testing.T) {
		var m *media.Media
		assert.False(t, m.IsImported())
	})
}

func TestMediaEnabled(t *testing.T) {
	t.Run("sets and gets enabled status", func(t *testing.T) {
		m := media.NewMedia("test")
		assert.False(t, m.IsEnabled())
		m.SetEnabled(true)
		assert.True(t, m.IsEnabled())
	})
}

func TestMediaIsSufficentlyProcessed(t *testing.T) {
	t.Run("returns false when no files", func(t *testing.T) {
		m := media.NewMedia("test")
		assert.False(t, m.IsSufficentlyProcessed(false))
	})
}

func TestCheckMediaQuality(t *testing.T) {
	tests := []struct {
		name    string
		quality string
		want    media.Quality
		valid   bool
	}{
		{"thumbnail is valid", "thumbnail", media.LowRes, true},
		{"fullres is valid", "fullres", media.HighRes, true},
		{"video is valid", "video", media.Video, true},
		{"invalid quality", "invalid", "", false},
		{"empty quality", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := media.CheckMediaQuality(tt.quality)
			assert.Equal(t, tt.valid, ok)

			if tt.valid {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestQualityConstants(t *testing.T) {
	t.Run("LowRes constant", func(t *testing.T) {
		assert.Equal(t, media.Quality("thumbnail"), media.LowRes)
	})

	t.Run("HighRes constant", func(t *testing.T) {
		assert.Equal(t, media.Quality("fullres"), media.HighRes)
	})

	t.Run("Video constant", func(t *testing.T) {
		assert.Equal(t, media.Quality("video"), media.Video)
	})
}

func TestSaveAndGetMedia(t *testing.T) {
	ctx := db.SetupTestDB(t, media.MediaCollectionKey)

	t.Run("saves and retrieves media", func(t *testing.T) {
		m := media.NewMedia("test-content-123")
		m.SetOwner("testuser")
		m.SetCreateDate(time.Now())
		m.SetEnabled(true)

		err := media.SaveMedia(ctx, m)
		require.NoError(t, err)

		retrieved, err := media.GetMediaByContentID(ctx, "test-content-123")
		require.NoError(t, err)
		assert.Equal(t, "test-content-123", retrieved.ID())
		assert.Equal(t, "testuser", retrieved.GetOwner())
	})

	t.Run("returns error for non-existent media", func(t *testing.T) {
		_, err := media.GetMediaByContentID(ctx, "non-existent")
		assert.Error(t, err)
	})
}

func TestMediaErrors(t *testing.T) {
	t.Run("ErrMediaNotFound", func(t *testing.T) {
		assert.Equal(t, "media not found", media.ErrMediaNotFound.Error())
	})

	t.Run("ErrMediaAlreadyExists", func(t *testing.T) {
		assert.Equal(t, "media already exists", media.ErrMediaAlreadyExists.Error())
	})

	t.Run("ErrNotDisplayable", func(t *testing.T) {
		assert.Equal(t, "media is not displayable", media.ErrNotDisplayable.Error())
	})

	t.Run("ErrMediaBadMimeType", func(t *testing.T) {
		assert.Equal(t, "media has a bad mime type", media.ErrMediaBadMimeType.Error())
	})
}
