package file_test

// import (
// 	"context"
// 	"crypto/rand"
// 	"encoding/base64"
// 	"math/big"
// 	"os"
// 	"testing"
// 	"time"
//
// 	"github.com/ethanrous/weblens/fileTree"
// 	"github.com/ethanrous/weblens/models"
// 	"github.com/ethanrous/weblens/modules/log"
// 	"github.com/rs/zerolog"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )
//
// func NewTestFileTree() (fileTree.FileTree, error) {
// 	logger := log.NewZeroLogger()
//
// 	hasher := mock.NewMockHasher()
// 	hasher.SetShouldCount(true)
// 	journal := mock.NewHollowJournalService()
//
// 	rootPath, err := os.MkdirTemp("", "weblens-test-*")
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	// MkdirTemp does not add a trailing slash to directories, which the fileTree expects
// 	rootPath += "/"
//
// 	logger.Trace().Func(func(e *zerolog.Event) { e.Msgf("Creating tmp root for FileTree test [%s]", rootPath) })
//
// 	tree, err := fileTree.NewFileTree(rootPath, "USERS", journal, false, logger)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return tree, nil
// }
//
// func NewTestFileService(name string, logger zerolog.Logger) (*models.ServicePack, error) {
// 	mondb, err := database.ConnectToMongo(env.GetMongoURI(), env.GetMongoDBName(env.Config{}), logger)
// 	if err != nil {
// 		return nil, err
// 	}
// 	usersCol := mondb.Collection(name + "users")
// 	err = usersCol.Drop(context.Background())
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer tests.CheckDropCol(usersCol, logger)
//
// 	accessCol := mondb.Collection(name + "access")
// 	err = accessCol.Drop(context.Background())
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer tests.CheckDropCol(accessCol, logger)
//
// 	folderMediaCol := mondb.Collection(name + "folderMedia")
// 	err = folderMediaCol.Drop(context.Background())
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer tests.CheckDropCol(folderMediaCol, logger)
//
// 	journalCol := mondb.Collection(name + "journal")
// 	err = journalCol.Drop(context.Background())
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer tests.CheckDropCol(journalCol, logger)
//
// 	serversCol := mondb.Collection(name + "servers")
// 	err = serversCol.Drop(context.Background())
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer tests.CheckDropCol(serversCol, logger)
//
// 	// Create the users tree
// 	usersTree, err := NewTestFileTree()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	// Create hasher and journal, and set the journal on the users tree
// 	hasherFactory := func() fileTree.Hasher {
// 		hasher := mock.NewMockHasher()
// 		hasher.SetShouldCount(true)
// 		return hasher
//
// 	}
//
// 	journalConfig := fileTree.JournalConfig{
// 		Collection:    journalCol,
// 		ServerID:      "TEST-SERVER",
// 		IgnoreLocal:   false,
// 		HasherFactory: hasherFactory,
// 		Logger:        logger,
// 	}
// 	journal, err := fileTree.NewJournal(journalConfig)
// 	if err != nil {
// 		return nil, err
// 	}
// 	usersTree.SetJournal(journal)
//
// 	// Create the cache and restore trees, and set their root aliases respectively
// 	cacheTree, err := NewTestFileTree()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	err = cacheTree.SetRootAlias(service.CachesTreeKey)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	restoreTree, err := NewTestFileTree()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	err = restoreTree.SetRootAlias(service.RestoreTreeKey)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	// Create user service
// 	userService, err := service.NewUserService(usersCol)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	// Create access service
// 	accessService, err := service.NewAccessService(userService, accessCol)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	// Create instance service
// 	instanceService, err := service.NewInstanceService(serversCol, logger)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	mediaService := &mock.MockMediaService{}
//
// 	fileService, err := service.NewFileService(
// 		logger, instanceService, userService, accessService, mediaService, folderMediaCol, usersTree, cacheTree, restoreTree,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return &models.ServicePack{
// 		FileService:     fileService,
// 		UserService:     userService,
// 		AccessService:   accessService,
// 		InstanceService: instanceService,
// 		MediaService:    mediaService,
// 	}, nil
// }
//
// // TestFileService_Restore_SingleFile tests the RestoreFiles method of the FileService on a single file
// func TestFileService_Restore_SingleFile(t *testing.T) {
// 	t.Parallel()
//
// 	logger := log.NewZeroLogger()
//
// 	pack, err := NewTestFileService(t.Name(), logger)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	pack.Caster = &mock.MockCaster{}
//
// 	// Create a user
// 	userName := "test-user"
// 	fullName := "Test User"
// 	password := "test-pass"
// 	testUser, err := models.NewUser(userName, password, fullName, false, true)
// 	require.NoError(t, err)
//
// 	usersTree, err := pack.FileService.GetFileTreeByName(service.UsersTreeKey)
// 	require.NoError(t, err)
//
// 	usersJournal := pack.FileService.GetJournalByTree(service.UsersTreeKey)
//
// 	setupEvent := usersJournal.NewEvent()
//
// 	// Create user home folder
// 	userHome, err := pack.FileService.CreateFolder(usersTree.GetRoot(), userName, setupEvent, pack.Caster)
// 	require.NoError(t, err)
//
// 	// Create user trash folder
// 	userTrash, err := pack.FileService.CreateFolder(userHome, service.UserTrashDirName, setupEvent, pack.Caster)
// 	require.NoError(t, err)
//
// 	testUser.SetHomeFolder(userHome)
// 	testUser.SetTrashFolder(userTrash)
//
// 	err = pack.UserService.Add(testUser)
// 	require.NoError(t, err)
//
// 	usersJournal.LogEvent(setupEvent)
// 	require.NoError(t, setupEvent.Wait())
//
// 	newFileEvent := usersJournal.NewEvent()
//
// 	// Get some data to write to the file
// 	fileSize, err := GenerateRandomInt(4096)
// 	require.NoError(t, err)
//
// 	testData, err := GenerateRandomBytes(fileSize)
// 	require.NoError(t, err)
//
// 	// Create a file
// 	testFileName := "test-file"
// 	testF, err := pack.FileService.CreateFile(userHome, testFileName, newFileEvent, pack.Caster, testData)
// 	require.NoError(t, err)
//
// 	// Commit the event
// 	usersJournal.LogEvent(newFileEvent)
// 	require.NoError(t, newFileEvent.Wait())
//
// 	beforeDelete := time.Now()
//
// 	// Move the file to the trash
// 	err = pack.FileService.MoveFilesToTrash([]*fileTree.WeblensFileImpl{testF}, testUser, nil, pack.Caster)
// 	require.NoError(t, err)
//
// 	err = pack.FileService.DeleteFiles([]*fileTree.WeblensFileImpl{testF}, service.UsersTreeKey, pack.Caster)
// 	require.NoError(t, err)
//
// 	// Restore the file
// 	err = pack.FileService.RestoreFiles([]string{testF.ID()}, userHome, beforeDelete, pack.Caster)
// 	require.NoError(t, err)
//
// 	// Check if the file is in the user's home
// 	restoredFile, err := userHome.GetChild(testFileName)
// 	if !assert.NoError(t, err) {
// 		logger.Error().Stack().Err(err).Msg("")
// 		t.FailNow()
// 	}
//
// 	assert.Equal(t, testF.GetContentID(), restoredFile.GetContentID())
//
// 	sz := restoredFile.Size()
// 	assert.Equal(t, int(sz), len(testData))
// }
//
// // TestFileService_Restore_Directory tests the RestoreFiles method of the FileService on a directory with sub-files
// func TestFileService_Restore_Directory(t *testing.T) {
// 	t.Parallel()
//
// 	logger := log.NewZeroLogger()
//
// 	pack, err := NewTestFileService(t.Name(), logger)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	pack.Caster = &mock.MockCaster{}
//
// 	uniqueUserExt, err := GenerateRandomBytes(8)
// 	require.NoError(t, err)
//
// 	// Create a user
// 	userName := "test-user" + base64.URLEncoding.EncodeToString(uniqueUserExt)
// 	fullName := "Test User"
// 	password := "test-pass"
// 	testUser, err := models.NewUser(userName, password, fullName, false, true)
// 	require.NoError(t, err)
//
// 	err = pack.FileService.CreateUserHome(testUser)
// 	require.NoError(t, err)
//
// 	err = pack.UserService.Add(testUser)
// 	require.NoError(t, err)
//
// 	userHome, err := pack.FileService.GetFileByTree(testUser.HomeID, service.UsersTreeKey)
// 	require.NoError(t, err)
//
// 	usersJournal := pack.FileService.GetJournalByTree(service.UsersTreeKey)
// 	event := usersJournal.NewEvent()
//
// 	// Create a random filesystem
// 	err = generateRandomFilesystem(pack.FileService, userHome, event, logger)
// 	require.NoError(t, err)
//
// 	// Commit the event to the journal
// 	usersJournal.LogEvent(event)
// 	// Wait for the event to be processed before continuing
// 	require.NoError(t, event.Wait())
//
// 	origHomeDirSize := userHome.Size()
// 	assert.NotEqual(t, origHomeDirSize, int64(0))
//
// 	allMyData, err := userHome.GetChild("literally-all-my-data")
// 	if !assert.NoError(t, err) {
// 		logger.Error().Stack().Err(err).Msg("")
// 		t.FailNow()
// 	}
//
// 	beforeDelete := time.Now()
//
// 	// Move the top directory to the trash
// 	err = pack.FileService.MoveFilesToTrash([]*fileTree.WeblensFileImpl{allMyData}, testUser, nil, pack.Caster)
// 	require.NoError(t, err)
//
// 	// Delete the top directory
// 	err = pack.FileService.DeleteFiles([]*fileTree.WeblensFileImpl{allMyData}, "USERS", pack.Caster)
// 	require.NoError(t, err)
//
// 	// Make sure home directory is empty
// 	assert.Equal(t, userHome.Size(), int64(0))
//
// 	// Restore the file
// 	err = pack.FileService.RestoreFiles([]string{allMyData.ID()}, userHome, beforeDelete, pack.Caster)
// 	if !assert.NoError(t, err) {
// 		logger.Error().Stack().Err(err).Msg("")
// 		t.FailNow()
// 	}
//
// 	// Check if the file is in the user's home
// 	_, err = userHome.GetChild("literally-all-my-data")
// 	if !assert.NoError(t, err) {
// 		logger.Error().Stack().Err(err).Msg("")
// 		t.FailNow()
// 	}
//
// 	// Make sure all the bytes came back
// 	postRestoreHomeDirSize := userHome.Size()
// 	assert.Equal(t, postRestoreHomeDirSize, origHomeDirSize)
// }
//
// func TestFileService_RestoreHistory(t *testing.T) {
// 	t.Parallel()
//
// 	logger := log.NewZeroLogger()
//
// 	pack, err := NewTestFileService(t.Name(), logger)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	pack.Caster = &mock.MockCaster{}
//
// 	usersTree, err := pack.FileService.GetFileTreeByName("USERS")
// 	require.NoError(t, err)
//
// 	usersJournal := pack.FileService.GetJournalByTree("USERS")
// 	if usersJournal == nil {
// 		t.Fatal("users journal not found")
// 	}
//
// 	event := usersJournal.NewEvent()
//
// 	filesCount, err := GenerateRandomInt(200)
// 	require.NoError(t, err)
//
// 	folders := []*fileTree.WeblensFileImpl{usersTree.GetRoot()}
// 	for range filesCount + 1 {
// 		parentIndex, err := GenerateRandomInt(len(folders))
// 		require.NoError(t, err)
//
// 		parent := folders[parentIndex]
// 		isFolder, err := GenerateRandomInt(2)
// 		require.NoError(t, err)
//
// 		b, err := GenerateRandomBytes(16)
// 		name := base64.URLEncoding.EncodeToString(b)
//
// 		require.NoError(t, err)
// 		if isFolder == 0 {
// 			logger.Trace().Func(func(e *zerolog.Event) { e.Msgf("Creating file [%s] with parent [%s]", name, parent.Filename()) })
//
// 			fileSize, err := GenerateRandomInt(4096) // Simulate writing data
// 			require.NoError(t, err)
//
// 			fileContent, err := GenerateRandomBytes(fileSize)
// 			require.NoError(t, err)
//
// 			_, err = pack.FileService.CreateFile(parent, name, event, pack.Caster, fileContent)
// 			if err != nil {
// 				t.Fatal(err)
// 			}
// 		} else {
// 			logger.Trace().Func(func(e *zerolog.Event) { e.Msgf("Creating folder [%s] with parent [%s]", name, parent.Filename()) })
// 			_, err = pack.FileService.CreateFolder(parent, name, event, pack.Caster)
// 			if err != nil {
// 				t.Fatal(err)
// 			}
// 		}
// 	}
//
// 	usersJournal.LogEvent(event)
// 	require.NoError(t, event.Wait())
//
// 	lifetimes := usersJournal.GetAllLifetimes()
// 	assert.Equal(t, filesCount+1, len(lifetimes))
//
// 	restorePack, err := NewTestFileService(t.Name()+"-restore", logger)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	err = restorePack.FileService.RestoreHistory(lifetimes)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	restoredUsersJournal := restorePack.FileService.GetJournalByTree("USERS")
// 	if restoredUsersJournal == nil {
// 		t.Fatal("restored users journal not found")
// 	}
//
// 	// TODO: NOT a good check. make more accurate to verify all files are accounted for
// 	assert.Equal(t, len(restoredUsersJournal.GetAllLifetimes()), len(lifetimes))
// }
//
// func GenerateRandomBytes(n int) ([]byte, error) {
// 	b := make([]byte, n)
// 	_, err := rand.Read(b)
// 	// Note that err == nil only if we read len(b) bytes.
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return b, nil
// }
//
// func GenerateRandomInt(max int) (int, error) {
// 	if max < 2 {
// 		return 0, nil
// 	}
//
// 	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
// 	if err != nil {
// 		return 0, err
// 	}
//
// 	return int(nBig.Int64()), nil
// }
//
// func generateRandomFilesystem(fileService models.FileService, rootFile *fileTree.WeblensFileImpl, event *fileTree.FileEvent, logger zerolog.Logger) error {
//
// 	caster := &mock.MockCaster{}
//
// 	subRootFolder, err := fileService.CreateFolder(rootFile, "literally-all-my-data", event, caster)
// 	if err != nil {
// 		return err
// 	}
//
// 	folders := []*fileTree.WeblensFileImpl{subRootFolder}
// 	fileCount, err := GenerateRandomInt(2048)
// 	if err != nil {
// 		return err
// 	}
//
// 	logger.Debug().Msgf("Creating %d random files", fileCount)
// 	for range fileCount {
// 		parentIndex, err := GenerateRandomInt(len(folders))
// 		if err != nil {
// 			return err
// 		}
// 		parent := folders[parentIndex]
//
// 		fType, err := GenerateRandomInt(2)
// 		if err != nil {
// 			return err
// 		}
//
// 		b, err := GenerateRandomBytes((parentIndex % 16) + 4)
// 		if err != nil {
// 			return err
// 		}
// 		filename := base64.URLEncoding.EncodeToString(b)
//
// 		if fType == 0 {
// 			// Create a folder
// 			newF, err := fileService.CreateFolder(parent, filename, event, caster)
// 			if err != nil {
// 				return err
// 			}
// 			folders = append(folders, newF)
// 		} else {
// 			// Create a file
// 			testF, err := fileService.CreateFile(parent, filename, event, caster)
// 			if err != nil {
// 				return err
// 			}
//
// 			// Write some data to the file
// 			testData, err := GenerateRandomBytes(4096)
// 			if err != nil {
// 				return err
// 			}
//
// 			_, err = testF.Write(testData)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}
//
// 	return fileService.ResizeDown(rootFile, event, caster)
// }
