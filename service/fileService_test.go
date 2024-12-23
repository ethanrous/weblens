package service_test

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ethanrous/weblens/database"
	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal/env"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/service"
	"github.com/ethanrous/weblens/service/mock"
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

	log.Trace.Func(func(l log.Logger) { l.Printf("Creating tmp root for FileTree test [%s]", rootPath) })

	tree, err := fileTree.NewFileTree(rootPath, "USERS", journal, false)
	if err != nil {
		return nil, err
	}

	return tree, nil
}

func NewTestFileService(name string, logger log.Bundle) (*models.ServicePack, error) {
	mondb, err := database.ConnectToMongo(env.GetMongoURI(), env.GetMongoDBName())
	if err != nil {
		return nil, err
	}
	usersCol := mondb.Collection(name + "users")
	err = usersCol.Drop(context.Background())
	if err != nil {
		return nil, err
	}
	defer usersCol.Drop(context.Background())

	accessCol := mondb.Collection(name + "access")
	err = accessCol.Drop(context.Background())
	if err != nil {
		return nil, err
	}
	defer accessCol.Drop(context.Background())

	folderMediaCol := mondb.Collection(name + "folderMedia")
	err = folderMediaCol.Drop(context.Background())
	if err != nil {
		return nil, err
	}
	defer folderMediaCol.Drop(context.Background())

	journalCol := mondb.Collection(name + "journal")
	err = journalCol.Drop(context.Background())
	if err != nil {
		return nil, err
	}
	defer journalCol.Drop(context.Background())

	serversCol := mondb.Collection(name + "servers")
	err = serversCol.Drop(context.Background())
	if err != nil {
		return nil, err
	}
	defer serversCol.Drop(context.Background())

	// Create the users tree
	usersTree, err := NewTestFileTree()
	if err != nil {
		return nil, err
	}

	// Create hasher and journal, and set the journal on the users tree
	hasherFactory := func() fileTree.Hasher {
		hasher := mock.NewMockHasher()
		hasher.SetShouldCount(true)
		return hasher

	}
	journal, err := fileTree.NewJournal(journalCol, "TEST-SERVER", false, hasherFactory, logger)
	if err != nil {
		return nil, err
	}
	usersTree.SetJournal(journal)

	// Create the cache and restore trees, and set their root aliases respectively
	cacheTree, err := NewTestFileTree()
	if err != nil {
		return nil, err
	}

	err = cacheTree.SetRootAlias("CACHES")
	if err != nil {
		return nil, err
	}

	restoreTree, err := NewTestFileTree()
	if err != nil {
		return nil, err
	}

	err = restoreTree.SetRootAlias("RESTORE")
	if err != nil {
		return nil, err
	}

	// Create user service
	userService, err := service.NewUserService(usersCol)
	if err != nil {
		return nil, err
	}

	// Create access service
	accessService, err := service.NewAccessService(userService, accessCol)
	if err != nil {
		return nil, err
	}

	// Create instance service
	instanceService, err := service.NewInstanceService(serversCol)
	if err != nil {
		return nil, err
	}

	mediaService := &mock.MockMediaService{}

	fileService, err := service.NewFileService(
		logger, instanceService, userService, accessService, mediaService, folderMediaCol, usersTree, cacheTree, restoreTree,
	)

	return &models.ServicePack{
		FileService:     fileService,
		UserService:     userService,
		AccessService:   accessService,
		InstanceService: instanceService,
		MediaService:    mediaService,
	}, nil
}

// TestFileService_Restore_SingleFile tests the RestoreFiles method of the FileService on a single file
func TestFileService_Restore_SingleFile(t *testing.T) {
	t.Parallel()

	logger := log.NewLogPackage("", log.DEBUG)

	pack, err := NewTestFileService(t.Name(), logger)
	if err != nil {
		t.Fatal(err)
	}

	pack.Caster = &mock.MockCaster{}

	// Create a user
	userName := "test-user"
	testUser, err := models.NewUser(userName, "test-pass", false, true)
	require.NoError(t, err)

	usersTree := pack.FileService.GetFileTreeByName("USERS")
	if usersTree == nil {
		t.Fatal("users tree not found")
	}

	usersJournal := pack.FileService.GetJournalByTree("USERS")

	setupEvent := usersJournal.NewEvent()

	// Create user home folder
	userHome, err := pack.FileService.CreateFolder(usersTree.GetRoot(), userName, setupEvent, pack.Caster)
	require.NoError(t, err)

	// Create user trash folder
	userTrash, err := pack.FileService.CreateFolder(userHome, ".user_trash", setupEvent, pack.Caster)
	require.NoError(t, err)

	testUser.SetHomeFolder(userHome)
	testUser.SetTrashFolder(userTrash)

	err = pack.UserService.Add(testUser)
	require.NoError(t, err)

	usersJournal.LogEvent(setupEvent)
	setupEvent.Wait()

	event := usersJournal.NewEvent()

	// Create a file
	testFileName := "test-file"
	testF, err := pack.FileService.CreateFile(userHome, testFileName, event, pack.Caster)
	require.NoError(t, err)

	// Write some data to the file
	testData := []byte("test")
	_, err = testF.Write(testData)
	require.NoError(t, err)

	// Commit the event
	usersJournal.LogEvent(event)

	event.Wait()

	beforeDelete := time.Now()

	// Move the file to the trash
	err = pack.FileService.MoveFilesToTrash([]*fileTree.WeblensFileImpl{testF}, testUser, nil, pack.Caster)
	require.NoError(t, err)

	err = pack.FileService.DeleteFiles([]*fileTree.WeblensFileImpl{testF}, "USERS", pack.Caster)
	require.NoError(t, err)

	// Restore the file
	err = pack.FileService.RestoreFiles([]string{testF.ID()}, userHome, beforeDelete, pack.Caster)
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

// TestFileService_Restore_Directory tests the RestoreFiles method of the FileService on a directory with sub-files
func TestFileService_Restore_Directory(t *testing.T) {
	t.Parallel()

	logger := log.NewLogPackage("", log.DEBUG)

	pack, err := NewTestFileService(t.Name(), logger)
	if err != nil {
		t.Fatal(err)
	}

	pack.Caster = &mock.MockCaster{}

	// Create a user
	userName := "test-user"
	testUser, err := models.NewUser(userName, "test-pass", false, true)
	require.NoError(t, err)

	err = pack.FileService.CreateUserHome(testUser)
	require.NoError(t, err)

	err = pack.UserService.Add(testUser)
	require.NoError(t, err)

	userHome, err := pack.FileService.GetFileByTree(testUser.HomeId, "USERS")
	require.NoError(t, err)

	usersJournal := pack.FileService.GetJournalByTree("USERS")
	event := usersJournal.NewEvent()

	// Create a directory
	testDirName := "test-dir"
	dir, err := pack.FileService.CreateFolder(userHome, testDirName, event, pack.Caster)
	require.NoError(t, err)

	testFileName := "test-file"
	fileCount := rand.Intn(20) + 1
	// Create 1-20 files in the directory
	for i := range fileCount {
		// Create a file
		testF, err := pack.FileService.CreateFile(dir, testFileName+strconv.Itoa(i), event, pack.Caster)
		require.NoError(t, err)

		// Write some data to the file
		testData := []byte("test" + strconv.Itoa(i))
		_, err = testF.Write(testData)
		require.NoError(t, err)

	}

	require.Equal(t, fileCount, len(dir.GetChildren()))

	// Commit the event
	usersJournal.LogEvent(event)

	event.Wait()

	beforeDelete := time.Now()

	// Move the file to the trash
	err = pack.FileService.MoveFilesToTrash([]*fileTree.WeblensFileImpl{dir}, testUser, nil, pack.Caster)
	require.NoError(t, err)

	err = pack.FileService.DeleteFiles([]*fileTree.WeblensFileImpl{dir}, "USERS", pack.Caster)
	require.NoError(t, err)

	// Add one here to account for the ROOT directory
	// require.Equal(t, fileCount+1, restoreTree.Size())

	// Restore the file
	err = pack.FileService.RestoreFiles([]string{dir.ID()}, userHome, beforeDelete, pack.Caster)
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

	// TODO: This fails when the random file count is 1
	assert.Equal(t, fileCount, len(restoredFile.GetChildren()))
}

func generateRandomString(n int) string {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num := rand.Intn(len(letters))
		ret[i] = letters[num]
	}

	return string(ret)
}

func TestFileService_RestoreHistory(t *testing.T) {
	t.Parallel()

	logger := log.NewLogPackage("", log.DEBUG)

	pack, err := NewTestFileService(t.Name(), logger)
	if err != nil {
		t.Fatal(err)
	}
	pack.Caster = &mock.MockCaster{}

	usersTree := pack.FileService.GetFileTreeByName("USERS")
	if usersTree == nil {
		t.Fatal("users tree not found")
	}

	usersJournal := pack.FileService.GetJournalByTree("USERS")
	if usersJournal == nil {
		t.Fatal("users journal not found")
	}

	event := usersJournal.NewEvent()

	filesCount := rand.Intn(200)
	folders := []*fileTree.WeblensFileImpl{usersTree.GetRoot()}
	for range filesCount {
		parent := folders[rand.Intn(len(folders))]
		isFolder := rand.Intn(2)
		name := generateRandomString(10)
		if isFolder == 0 {
			log.Trace.Func(func(l log.Logger) { l.Printf("Creating file [%s] with parent [%s]", name, parent.Filename()) })
			_, err = pack.FileService.CreateFile(parent, name, event, pack.Caster)
			if err != nil {
				t.Fatal(err)
			}
		} else {
			log.Trace.Func(func(l log.Logger) { l.Printf("Creating folder [%s] with parent [%s]", name, parent.Filename()) })
			_, err = pack.FileService.CreateFolder(parent, name, event, pack.Caster)
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	usersJournal.LogEvent(event)
	event.Wait()

	lifetimes := usersJournal.GetAllLifetimes()
	assert.Equal(t, filesCount, len(lifetimes))

	restorePack, err := NewTestFileService(t.Name()+"-restore", logger)
	if err != nil {
		t.Fatal(err)
	}

	err = restorePack.FileService.RestoreHistory(lifetimes)
	if err != nil {
		t.Fatal(err)
	}

	restoredUsersJournal := restorePack.FileService.GetJournalByTree("USERS")
	if restoredUsersJournal == nil {
		t.Fatal("restored users journal not found")
	}

	assert.Equal(t, len(restoredUsersJournal.GetAllLifetimes()), len(lifetimes))
}
