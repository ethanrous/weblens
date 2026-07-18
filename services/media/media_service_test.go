package media_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"

	file_model "github.com/ethanrous/weblens/models/file"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/modules/wlfs"
	"github.com/ethanrous/weblens/modules/wlog"
	"github.com/ethanrous/weblens/services/ctxservice"
	file_service "github.com/ethanrous/weblens/services/file"
	media_service "github.com/ethanrous/weblens/services/media"
	"github.com/stretchr/testify/require"
	"github.com/viccon/sturdyc"
)

const testImageFixture = "../../images/testMedia/DSC08113.jpg"

// newMediaTestContext creates an AppContext with a real FileService over a temp
// filesystem, and returns the context along with the absolute thumbs directory.
func newMediaTestContext(t *testing.T) (ctxservice.AppContext, string) {
	t.Helper()

	tempDir := t.TempDir()
	usersDir := filepath.Join(tempDir, "USERS")
	cachesDir := filepath.Join(tempDir, "CACHES")
	thumbsDir := filepath.Join(cachesDir, file_model.ThumbsDirName)

	require.NoError(t, wlfs.RegisterAbsolutePrefix(file_model.UsersTreeKey, usersDir))
	require.NoError(t, wlfs.RegisterAbsolutePrefix(file_model.CachesTreeKey, cachesDir))
	require.NoError(t, os.MkdirAll(thumbsDir, 0755))

	logger := wlog.NewZeroLogger()
	basicCtx := ctxservice.NewBasicContext(context.Background(), logger)

	fsSvc, err := file_service.NewFileService(basicCtx)
	require.NoError(t, err)

	appCtx := ctxservice.AppContext{
		BasicContext: basicCtx,
		FileService:  fsSvc,
		Cache:        make(map[string]*sturdyc.Client[any]),
		WG:           &sync.WaitGroup{},
	}

	return appCtx, thumbsDir
}

// copyFixtureToUsersTree copies the test image into the users tree and returns
// its WeblensFileImpl.
func copyFixtureToUsersTree(t *testing.T, filename string) *file_model.WeblensFileImpl {
	t.Helper()

	photoPath := file_model.UsersRootPath.Child("testuser", true).Child(filename, false)

	require.NoError(t, os.MkdirAll(filepath.Dir(photoPath.ToAbsolute()), 0755))

	src, err := os.Open(testImageFixture)
	require.NoError(t, err)

	defer src.Close() //nolint:errcheck

	dst, err := os.Create(photoPath.ToAbsolute())
	require.NoError(t, err)

	defer dst.Close() //nolint:errcheck

	_, err = io.Copy(dst, src)
	require.NoError(t, err)

	return file_model.NewWeblensFile(file_model.NewFileOptions{
		Path:      photoPath,
		ContentID: "testcontentid123",
	})
}

func TestImportMediaFromFile(t *testing.T) {
	appCtx, thumbsDir := newMediaTestContext(t)
	f := copyFixtureToUsersTree(t, "DSC08113.jpg")

	m, err := media_service.ImportMediaFromFile(appCtx, f)
	require.NoError(t, err)
	require.NotNil(t, m)

	require.Equal(t, media_model.ContentID("testcontentid123"), m.ID())
	require.Equal(t, 1, m.PageCount)
	require.Greater(t, m.Width, 0)
	require.Greater(t, m.Height, 0)

	// Both thumbnail qualities must exist on disk with real content
	for _, quality := range []media_model.Quality{media_model.LowRes, media_model.HighRes} {
		name, err := media_model.FmtCacheFileName(m.ID(), quality, 0)
		require.NoError(t, err)

		info, err := os.Stat(filepath.Join(thumbsDir, name))
		require.NoError(t, err, "expected %s cache file to exist", quality)
		require.Greater(t, info.Size(), int64(0), "%s cache file should not be empty", quality)
	}
}
