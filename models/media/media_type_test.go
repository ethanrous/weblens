package media_test

import (
	"testing"

	"github.com/ethanrous/weblens/models/media"
	"github.com/stretchr/testify/assert"
)

func TestParseExtension(t *testing.T) {
	tests := []struct {
		name          string
		ext           string
		expectMime    string
		isDisplayable bool
		isRaw         bool
		isVideo       bool
	}{
		{"jpeg extension", "jpg", "image/jpeg", true, false, false},
		{"jpeg with dot", ".jpg", "image/jpeg", true, false, false},
		{"png extension", "png", "image/png", true, false, false},
		{"gif extension", "gif", "image/gif", true, false, false},
		{"webp extension", "webp", "image/webp", true, false, false},
		{"nef raw extension", "NEF", "image/x-nikon-nef", true, true, false},
		{"arw raw extension", "ARW", "image/x-sony-arw", true, true, false},
		{"mp4 video", "mp4", "video/mp4", true, false, true},
		{"mkv video", "mkv", "video/x-matroska", true, false, true},
		{"pdf document", "pdf", "application/pdf", true, false, false},
		{"zip archive", "zip", "application/zip", false, false, false},
		{"unknown extension", "xyz", "generic", false, false, false},
		{"empty extension", "", "generic", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := media.ParseExtension(tt.ext)
			assert.Equal(t, tt.expectMime, mt.Mime)
			assert.Equal(t, tt.isDisplayable, mt.Displayable)
			assert.Equal(t, tt.isRaw, mt.Raw)
			assert.Equal(t, tt.isVideo, mt.IsVideo)
		})
	}
}

func TestParseMime(t *testing.T) {
	tests := []struct {
		name string
		mime string
		want string
	}{
		{"image/jpeg", "image/jpeg", "Jpeg"},
		{"image/png", "image/png", "Png"},
		{"image/gif", "image/gif", "Gif"},
		{"video/mp4", "video/mp4", "MP4"},
		{"application/pdf", "application/pdf", "PDF"},
		{"unknown mime", "unknown/type", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := media.ParseMime(tt.mime)
			assert.Equal(t, tt.want, mt.Name)
		})
	}
}

func TestGeneric(t *testing.T) {
	t.Run("returns generic media type", func(t *testing.T) {
		mt := media.Generic()
		assert.Equal(t, "generic", mt.Mime)
		assert.Equal(t, "File", mt.Name)
		assert.False(t, mt.Displayable)
	})
}

func TestSize(t *testing.T) {
	t.Run("returns number of media types", func(t *testing.T) {
		size := media.Size()
		assert.Greater(t, size, 0)
	})
}

func TestGetMaps(t *testing.T) {
	t.Run("returns both maps", func(t *testing.T) {
		mimeMap, extMap := media.GetMaps()
		assert.NotEmpty(t, mimeMap)
		assert.NotEmpty(t, extMap)
	})
}

func TestMTypeIsMime(t *testing.T) {
	t.Run("returns true for matching mime", func(t *testing.T) {
		mt := media.ParseMime("image/jpeg")
		assert.True(t, mt.IsMime("image/jpeg"))
	})

	t.Run("returns false for non-matching mime", func(t *testing.T) {
		mt := media.ParseMime("image/jpeg")
		assert.False(t, mt.IsMime("image/png"))
	})
}

func TestMTypeIsDisplayable(t *testing.T) {
	t.Run("jpeg is displayable", func(t *testing.T) {
		mt := media.ParseMime("image/jpeg")
		assert.True(t, mt.IsDisplayable())
	})

	t.Run("zip is not displayable", func(t *testing.T) {
		mt := media.ParseMime("application/zip")
		assert.False(t, mt.IsDisplayable())
	})
}

func TestMTypeFriendlyName(t *testing.T) {
	tests := []struct {
		mime     string
		expected string
	}{
		{"image/jpeg", "Jpeg"},
		{"image/png", "Png"},
		{"video/mp4", "MP4"},
		{"application/pdf", "PDF"},
	}

	for _, tt := range tests {
		t.Run(tt.mime, func(t *testing.T) {
			mt := media.ParseMime(tt.mime)
			assert.Equal(t, tt.expected, mt.FriendlyName())
		})
	}
}

func TestMTypeIsSupported(t *testing.T) {
	t.Run("jpeg is supported", func(t *testing.T) {
		mt := media.ParseMime("image/jpeg")
		assert.True(t, mt.IsSupported())
	})

	t.Run("generic is not supported", func(t *testing.T) {
		mt := media.Generic()
		assert.False(t, mt.IsSupported())
	})

	t.Run("unknown mime is not supported", func(t *testing.T) {
		mt := media.ParseMime("unknown/type")
		assert.False(t, mt.IsSupported())
	})
}

func TestMTypeIsMultiPage(t *testing.T) {
	t.Run("pdf is multi-page", func(t *testing.T) {
		mt := media.ParseMime("application/pdf")
		assert.True(t, mt.IsMultiPage())
	})

	t.Run("jpeg is not multi-page", func(t *testing.T) {
		mt := media.ParseMime("image/jpeg")
		assert.False(t, mt.IsMultiPage())
	})
}

func TestMTypeGetThumbExifKey(t *testing.T) {
	t.Run("nikon raw has JpgFromRaw key", func(t *testing.T) {
		mt := media.ParseMime("image/x-nikon-nef")
		assert.Equal(t, "JpgFromRaw", mt.GetThumbExifKey())
	})

	t.Run("sony raw has PreviewImage key", func(t *testing.T) {
		mt := media.ParseMime("image/x-sony-arw")
		assert.Equal(t, "PreviewImage", mt.GetThumbExifKey())
	})

	t.Run("jpeg has empty key", func(t *testing.T) {
		mt := media.ParseMime("image/jpeg")
		assert.Empty(t, mt.GetThumbExifKey())
	})
}

func TestMTypeSupportsImgRecog(t *testing.T) {
	t.Run("jpeg supports image recognition", func(t *testing.T) {
		mt := media.ParseMime("image/jpeg")
		assert.True(t, mt.SupportsImgRecog())
	})

	t.Run("gif does not support image recognition", func(t *testing.T) {
		mt := media.ParseMime("image/gif")
		assert.False(t, mt.SupportsImgRecog())
	})

	t.Run("video does not support image recognition", func(t *testing.T) {
		mt := media.ParseMime("video/mp4")
		assert.False(t, mt.SupportsImgRecog())
	})
}
