package service_test

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ethrousseau/weblens/database"
	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal/env"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/service"
	"github.com/ethrousseau/weblens/service/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func NewTestFileTree() (fileTree.FileTree, error) {
	hasher := mock.NewMockHasher()
	hasher.SetShouldCount(true)
	journal := mock.NewHollowJournalService()

	rootPath, err := os.MkdirTemp("", "weblens-test-*")
	if err != nil {
		return nil, err
	}

	// MkdirTemp does not add a trailing slash to directories, which the fileTree expects
	rootPath += "/"

	log.Trace.Printf("Creating tmp root for FileTree test [%s]", rootPath)

	tree, err := fileTree.NewFileTree(rootPath, "USERS", journal)
	if err != nil {
		return nil, err
	}

	return tree, nil
}

// TestFileService_Restore_SingleFile tests the RestoreFiles method of the FileService on a single file
func TestFileService_Restore_SingleFile(t *testing.T) {
	mondb, err := database.ConnectToMongo(env.GetMongoURI(), env.GetMongoDBName())
	require.NoError(t, err)
	usersCol := mondb.Collection(t.Name() + "users")
	usersCol.Drop(context.Background())
	defer usersCol.Drop(context.Background())

	accessCol := mondb.Collection(t.Name() + "access")
	accessCol.Drop(context.Background())
	defer accessCol.Drop(context.Background())

	trashCol := mondb.Collection(t.Name() + "trash")
	trashCol.Drop(context.Background())
	defer trashCol.Drop(context.Background())

	journalCol := mondb.Collection(t.Name() + "journal")
	journalCol.Drop(context.Background())
	defer journalCol.Drop(context.Background())

	// Create the users tree
	usersTree, err := NewTestFileTree()
	require.NoError(t, err)

	// Create hasher and journal, and set the journal on the users tree
	hasherFactory := func() fileTree.Hasher {
		hasher := mock.NewMockHasher()
		hasher.SetShouldCount(true)
		return hasher

	}
	journal, err := fileTree.NewJournal(journalCol, "TEST-SERVER", false, hasherFactory)
	require.NoError(t, err)
	usersTree.SetJournal(journal)

	// Create the cache and restore trees, and set their root aliases respectively
	cacheTree, err := NewTestFileTree()
	require.NoError(t, err)
	usersTree.SetRootAlias("CACHES")

	retoreTree, err := NewTestFileTree()
	require.NoError(t, err)
	usersTree.SetRootAlias("RESTORE")

	// Create user service
	userService, err := service.NewUserService(usersCol)
	require.NoError(t, err)

	// Create access service
	accessService, err := service.NewAccessService(userService, accessCol)
	require.NoError(t, err)

	mediaService := &mock.MockMediaService{}

	fileService, err := service.NewFileService(
		usersTree, cacheTree, retoreTree, userService, accessService, mediaService, trashCol,
	)
	require.NoError(t, err)

	caster := &mock.MockCaster{}

	// Create a user
	userName := "test-user"
	testUser, err := models.NewUser(userName, "test-pass", false, true)
	require.NoError(t, err)
	err = userService.Add(testUser)
	require.NoError(t, err)

	// Create user home folder
	userHome, err := fileService.CreateFolder(fileService.GetMediaRoot(), userName, caster)
	require.NoError(t, err)

	// Create user trash folder
	userTrash, err := fileService.CreateFolder(userHome, ".user_trash", caster)
	require.NoError(t, err)

	testUser.SetHomeFolder(userHome)
	testUser.SetTrashFolder(userTrash)

	event := fileService.GetUsersJournal().NewEvent()

	// Create a file
	testFileName := "test-file"
	testF, err := fileService.CreateFile(userHome, testFileName, event)
	require.NoError(t, err)

	// Write some data to the file
	testData := []byte("test")
	_, err = testF.Write(testData)
	require.NoError(t, err)

	// Commit the event
	fileService.GetUsersJournal().LogEvent(event)

	event.Wait()

	beforeDelete := time.Now()

	// Move the file to the trash
	err = fileService.MoveFilesToTrash([]*fileTree.WeblensFileImpl{testF}, testUser, nil, caster)
	require.NoError(t, err)

	err = fileService.DeleteFiles([]*fileTree.WeblensFileImpl{testF}, caster)
	require.NoError(t, err)

	// Restore the file
	err = fileService.RestoreFiles([]string{testF.ID()}, userHome, beforeDelete, caster)
	require.NoError(t, err)

	// Check if the file is in the user's home
	restoredFile, err := userHome.GetChild(testFileName)
	if !assert.NoError(t, err) {
		log.ErrTrace(err)
		t.FailNow()
	}

	assert.Equal(t, testF.GetContentId(), restoredFile.GetContentId())

	sz := restoredFile.Size()
	assert.Equal(t, int(sz), len(testData))
}

// TestFileService_Restore_SingleFile tests the RestoreFiles method of the FileService on a directory with sub-files
func TestFileService_Restore_Directory(t *testing.T) {
	mondb, err := database.ConnectToMongo(env.GetMongoURI(), env.GetMongoDBName())
	require.NoError(t, err)
	usersCol := mondb.Collection(t.Name() + "users")
	usersCol.Drop(context.Background())
	defer usersCol.Drop(context.Background())

	accessCol := mondb.Collection(t.Name() + "access")
	accessCol.Drop(context.Background())
	defer accessCol.Drop(context.Background())

	trashCol := mondb.Collection(t.Name() + "trash")
	trashCol.Drop(context.Background())
	defer trashCol.Drop(context.Background())

	journalCol := mondb.Collection(t.Name() + "journal")
	journalCol.Drop(context.Background())
	defer journalCol.Drop(context.Background())

	usersTree, err := NewTestFileTree()
	require.NoError(t, err)
	usersTree.SetRootAlias("USERS")

	hasherFactory := func() fileTree.Hasher {
		hasher := mock.NewMockHasher()
		hasher.SetShouldCount(true)
		return hasher

	}
	journal, err := fileTree.NewJournal(journalCol, "TEST-SERVER", false, hasherFactory)
	require.NoError(t, err)
	usersTree.SetJournal(journal)

	cacheTree, err := NewTestFileTree()
	require.NoError(t, err)
	restoreTree, err := NewTestFileTree()
	require.NoError(t, err)

	userService, err := service.NewUserService(usersCol)
	require.NoError(t, err)

	accessService, err := service.NewAccessService(userService, accessCol)
	require.NoError(t, err)

	mediaService := &mock.MockMediaService{}

	fileService, err := service.NewFileService(
		usersTree, cacheTree, restoreTree, userService, accessService, mediaService, trashCol,
	)
	require.NoError(t, err)

	caster := &mock.MockCaster{}

	// Create a user
	userName := "test-user"
	testUser, err := models.NewUser(userName, "test-pass", false, true)
	require.NoError(t, err)
	err = userService.Add(testUser)
	require.NoError(t, err)

	// Create user home folder
	userHome, err := fileService.CreateFolder(fileService.GetMediaRoot(), userName, caster)
	require.NoError(t, err)

	// Create user trash folder
	userTrash, err := fileService.CreateFolder(userHome, ".user_trash", caster)
	require.NoError(t, err)

	testUser.SetHomeFolder(userHome)
	testUser.SetTrashFolder(userTrash)

	event := fileService.GetUsersJournal().NewEvent()

	// Create a directory
	testDirName := "test-dir"
	dir, err := fileService.CreateFolder(userHome, testDirName, caster)
	require.NoError(t, err)

	testFileName := "test-file"
	fileCount := rand.Intn(20)
	// Create up to 20 files in the directory
	for i := range fileCount {
		// Create a file
		testF, err := fileService.CreateFile(dir, testFileName+strconv.Itoa(i), event)
		require.NoError(t, err)

		// Write some data to the file
		testData := []byte("test" + strconv.Itoa(i))
		_, err = testF.Write(testData)
		require.NoError(t, err)

	}

	require.Equal(t, fileCount, len(dir.GetChildren()))

	// Commit the event
	fileService.GetUsersJournal().LogEvent(event)

	event.Wait()

	beforeDelete := time.Now()

	// Move the file to the trash
	err = fileService.MoveFilesToTrash([]*fileTree.WeblensFileImpl{dir}, testUser, nil, caster)
	require.NoError(t, err)

	err = fileService.DeleteFiles([]*fileTree.WeblensFileImpl{dir}, caster)
	require.NoError(t, err)

	// Add one here to account for the ROOT directory
	require.Equal(t, fileCount+1, restoreTree.Size())

	// Restore the file
	err = fileService.RestoreFiles([]string{dir.ID()}, userHome, beforeDelete, caster)
	if !assert.NoError(t, err) {
		log.ErrTrace(err)
		t.FailNow()
	}

	// Check if the file is in the user's home
	restoredFile, err := userHome.GetChild(testDirName)
	if !assert.NoError(t, err) {
		log.ErrTrace(err)
		t.FailNow()
	}

	assert.Equal(t, fileCount, len(restoredFile.GetChildren()))

}
