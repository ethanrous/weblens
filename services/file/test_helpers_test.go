package file //nolint:testpackage

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/ethanrous/weblens/models/db"
	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/history"
	media_model "github.com/ethanrous/weblens/models/media"
	tower_model "github.com/ethanrous/weblens/models/tower"
	"github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/services/ctxservice"
	"github.com/ethanrous/weblens/services/notify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/viccon/sturdyc"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// initializeTestRoots initializes the file service with root files for testing.
func (fsSvc *ServiceImpl) initializeTestRoots(_ context.Context, roots ...fs.Filepath) error {
	for _, rootPath := range roots {
		rootFile := file_model.NewWeblensFile(file_model.NewFileOptions{
			Path:       rootPath,
			CreateNow:  true,
			GenerateID: true,
		})
		if rootFile == nil {
			continue
		}

		// Add root file directly to the internal map, bypassing parent checks
		fsSvc.setFileInternal(rootFile.ID(), rootFile)
	}

	return nil
}

// integrationTestCleanup contains cleanup functions to run after test.
type integrationTestCleanup struct {
	tempDir string
}

// cleanup runs all cleanup tasks.
func (c *integrationTestCleanup) cleanup(t *testing.T) {
	t.Helper()

	if c.tempDir != "" {
		err := os.RemoveAll(c.tempDir)
		if err != nil {
			t.Logf("Failed to remove temp directory: %v", err)
		}
	}
}

// testContextOptions configures the integration test context.
type testContextOptions struct {
	towerRole tower_model.Role
	username  string
}

// testContextOption is a functional option for configuring integration test context.
type testContextOption func(*testContextOptions)

// withTowerRole sets the tower role for the test (default: RoleCore).
func withTowerRole(role tower_model.Role) testContextOption {
	return func(opts *testContextOptions) {
		opts.towerRole = role
	}
}

// newIntegrationTestContext creates a complete integration test context with:
// - Real MongoDB connection with test collections
// - Temporary filesystem with USERS/, RESTORE/, BACKUP/ structure
// - Tower document in database with specified role
// - Test user home and trash directories
// - Initialized FileService
// - AppContext with all services wired
// - FileEvent for history tracking
func newIntegrationTestContext(t *testing.T, opts ...testContextOption) (context.Context, *integrationTestCleanup) {
	t.Helper()

	// Apply options
	options := &testContextOptions{
		towerRole: tower_model.RoleCore,
		username:  "testuser",
	}
	for _, opt := range opts {
		opt(options)
	}

	// 1. Setup database with required collections
	dbCtx := db.SetupTestDB(t, history.FileHistoryCollectionKey)

	// Also setup media collection
	col, err := db.GetCollection[any](dbCtx, media_model.MediaCollectionKey)
	require.NoError(t, err)
	err = col.Drop(dbCtx)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = col.Drop(dbCtx)
	})

	// Setup tower collection
	towerCol, err := db.GetCollection[any](dbCtx, tower_model.TowerCollectionKey)
	require.NoError(t, err)
	err = towerCol.Drop(dbCtx)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = towerCol.Drop(dbCtx)
	})

	// 2. Create temp filesystem with required structure
	tempDir := t.TempDir()

	// Create required root directories
	usersDir := filepath.Join(tempDir, "USERS")
	restoreDir := filepath.Join(tempDir, "RESTORE")
	backupDir := filepath.Join(tempDir, "BACKUP")
	cachesDir := filepath.Join(tempDir, "CACHES")

	require.NoError(t, os.MkdirAll(usersDir, 0755))
	require.NoError(t, os.MkdirAll(restoreDir, 0755))
	require.NoError(t, os.MkdirAll(backupDir, 0755))
	require.NoError(t, os.MkdirAll(cachesDir, 0755))

	// Register the filesystem root paths for testing
	require.NoError(t, fs.RegisterAbsolutePrefix(file_model.UsersTreeKey, usersDir))
	require.NoError(t, fs.RegisterAbsolutePrefix(file_model.RestoreTreeKey, restoreDir))
	require.NoError(t, fs.RegisterAbsolutePrefix(file_model.BackupTreeKey, backupDir))
	require.NoError(t, fs.RegisterAbsolutePrefix(file_model.CachesTreeKey, cachesDir))

	// 3. Create tower document in database
	towerID := primitive.NewObjectID().Hex()
	tower := tower_model.Instance{
		TowerID:     towerID,
		Name:        "Test Tower",
		Role:        options.towerRole,
		IsThisTower: true,
		DbID:        primitive.NewObjectID(),
	}
	_, err = towerCol.GetCollection().InsertOne(dbCtx, tower)
	require.NoError(t, err)

	// 4. Note: User directories will be created via file service after it's initialized

	// 5. Create logger and basic context
	logger := log.NewZeroLogger()
	basicCtx := ctxservice.NewBasicContext(dbCtx, logger)

	// 6. Get database from context for AppContext
	dbAny := dbCtx.Value(db.DatabaseContextKey)
	database, _ := dbAny.(*mongo.Database)

	// 7. Create AppContext with LocalTowerID
	appCtx := ctxservice.AppContext{
		BasicContext: basicCtx,
		DB:           database,
		Cache:        make(map[string]*sturdyc.Client[any]),
		WG:           &sync.WaitGroup{},
		LocalTowerID: towerID,
	}

	// 8. Initialize ClientService (notification service)
	fileServiceCtx := appCtx.WithContext(dbCtx)
	appCtx.ClientService = notify.NewClientManager(fileServiceCtx)

	// 9. Initialize FileService
	fsSvc, err := NewFileService(fileServiceCtx)
	require.NoError(t, err)

	appCtx.FileService = fsSvc

	// 10. Initialize file tree roots using internal test helper
	err = fsSvc.initializeTestRoots(fileServiceCtx,
		file_model.UsersRootPath,
		file_model.RestoreDirPath,
		file_model.BackupRootPath,
		file_model.CacheRootPath,
	)
	require.NoError(t, err)

	// 11. Add FileEvent and towerID to context for history tracking
	ctx := appCtx.WithContext(dbCtx)
	ctx = context.WithValue(ctx, "towerID", towerID) //nolint:revive
	ctx = history.WithFileEvent(ctx)

	// 12. Create user home directory using file service
	usersRoot, err := fsSvc.GetFileByID(ctx, file_model.UsersTreeKey)
	require.NoError(t, err, "failed to get USERS root")

	userHome, err := fsSvc.CreateFolder(ctx, usersRoot, options.username)
	require.NoError(t, err, "failed to create user home directory")

	// Create .user_trash directory
	_, err = fsSvc.CreateFolder(ctx, userHome, file_model.UserTrashDirName)
	require.NoError(t, err, "failed to create user trash directory")

	cleanup := &integrationTestCleanup{
		tempDir: tempDir,
	}

	t.Cleanup(func() {
		cleanup.cleanup(t)
	})

	return ctx, cleanup
}

// createTestFile creates a file on disk and adds it to the file service.
// Returns the WeblensFileImpl representing the created file.
func createTestFile(t *testing.T, ctx context.Context, fs file_model.Service, parent *file_model.WeblensFileImpl, filename string, content []byte) *file_model.WeblensFileImpl { //nolint:revive
	t.Helper()

	// Create file in service
	file, err := fs.CreateFile(ctx, parent, filename, content)
	require.NoError(t, err)

	return file
}

// createTestFolder creates a folder on disk and adds it to the file service.
// Returns the WeblensFileImpl representing the created folder.
func createTestFolder(t *testing.T, ctx context.Context, fs file_model.Service, parent *file_model.WeblensFileImpl, folderName string) *file_model.WeblensFileImpl { //nolint:revive
	t.Helper()

	// Create folder in service (this should also create on disk)
	folder, err := fs.CreateFolder(ctx, parent, folderName)
	require.NoError(t, err)

	return folder
}

// assertFileExistsOnDisk verifies that a file exists on the filesystem.
func assertFileExistsOnDisk(t *testing.T, path fs.Filepath) {
	t.Helper()

	absPath := path.ToAbsolute()
	_, err := os.Stat(absPath)
	assert.NoError(t, err, "file should exist on disk: %s", absPath)
}

// assertFileNotExistsOnDisk verifies that a file does not exist on the filesystem.
func assertFileNotExistsOnDisk(t *testing.T, path fs.Filepath) {
	t.Helper()

	absPath := path.ToAbsolute()
	_, err := os.Stat(absPath)
	assert.True(t, os.IsNotExist(err), "file should not exist on disk: %s", absPath)
}

// assertFileNotInService verifies that a file does not exist in the file service.
func assertFileNotInService(t *testing.T, ctx context.Context, fs file_model.Service, fileID string) { //nolint:revive
	t.Helper()

	_, err := fs.GetFileByID(ctx, fileID)
	assert.Error(t, err, "file should not exist in service")
	assert.ErrorIs(t, err, file_model.ErrFileNotFound)
}
