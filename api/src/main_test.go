package main

import (
	"io/fs"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/ethrousseau/weblens/api/dataProcess"
	"github.com/ethrousseau/weblens/api/dataStore"
	"github.com/ethrousseau/weblens/api/dataStore/database"
	"github.com/ethrousseau/weblens/api/dataStore/filetree"
	"github.com/ethrousseau/weblens/api/dataStore/history"
	"github.com/ethrousseau/weblens/api/dataStore/instance"
	"github.com/ethrousseau/weblens/api/dataStore/media"
	"github.com/ethrousseau/weblens/api/dataStore/user"
	"github.com/ethrousseau/weblens/api/routes"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"github.com/stretchr/testify/assert"
)

const TestUserName = types.Username("gwen")
const TestUserPass = "stacy!23"

const MediaRootPath = "/Users/ethan/weblens/test/media"

func mainTestInit(t *testing.T) {
	if types.SERV.StoreService != nil {
		return
	}

	// Create + set test database controller
	db := database.New(util.GetMongoURI(), "weblens-test")
	types.SERV.SetStore(db)

	instanceService := instance.NewService()
	types.SERV.SetInstance(instanceService)

	// Create and set hollow caster, we are not testing websockets here
	caster := routes.NewCaster()
	caster.Disable()
	types.SERV.SetCaster(caster)

	// Cleanup any old/existing users
	err := db.DeleteAllUsers()
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}

	// Cleanup any old/existing media
	err = db.DeleteAllMedia()
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}

	// Create + init + set mock user service
	us := user.NewService()
	err = us.Init(db)
	types.SERV.SetUserService(us)
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}

	// Remove all files from real filesystem where file tree will be initialized
	err = os.RemoveAll(MediaRootPath)
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}

	// Create + set new test file service as "MEDIA"
	ft := filetree.NewFileTree(MediaRootPath, "MEDIA")
	types.SERV.SetFileTree(ft)

	js := history.NewService(ft, db)
	ft.SetJournal(js)

	// Create and add new user
	gwenUser, err := user.New(TestUserName, TestUserPass, false, true)
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}
	if !assert.NotNil(t, gwenUser) {
		t.Fatal("user should not be nil")
	}
	err = types.SERV.UserService.Add(gwenUser)
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}

	// TODO - create seperate file trees for seperate roots
	// Set caches path so tree init passes
	err = os.Setenv("CACHES_PATH", "/Users/ethan/weblens/test/cache")
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}

	// Path of static test media content in the repo
	testMediaSourcePath := "/Users/ethan/repos/weblens/api/testMedia"
	dirents, err := os.ReadDir(testMediaSourcePath)
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}

	for _, dirent := range dirents {
		if dirent.IsDir() {
			continue
		}
		sourceFile := testMediaSourcePath + "/" + dirent.Name()
		data, err := os.ReadFile(sourceFile)
		if !assert.NoError(t, err) {
			t.Fatal(err)
		}

		destFile := gwenUser.GetHomeFolder().GetAbsPath() + dirent.Name()
		err = os.WriteFile(destFile, data, fs.ModePerm)
		if !assert.NoError(t, err) {
			t.Fatal(err)
		}
	}

	mediaTypeServ := media.NewTypeService()
	mediaService := media.NewRepo(mediaTypeServ)
	err = mediaService.Init(db)
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}
	types.SERV.SetMediaRepo(mediaService)

	workerPool, taskDispatcher := dataProcess.NewWorkerPool(runtime.NumCPU() - 2)
	types.SERV.SetWorkerPool(workerPool)
	types.SERV.SetTaskDispatcher(taskDispatcher)
	workerPool.Run()

	err = dataStore.InitMediaRoot(types.SERV.FileTree)
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}

}

func mainTestCleanup(t *testing.T) {
	err := types.SERV.StoreService.DeleteAllUsers()
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}

	// Cleanup any old/existing media
	err = types.SERV.StoreService.DeleteAllMedia()
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}

	err = types.SERV.StoreService.DeleteAllFileHistory()
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}

	err = os.RemoveAll(types.SERV.FileTree.GetRoot().GetAbsPath())
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}

	err = os.RemoveAll(types.SERV.FileTree.Get("CACHE").GetAbsPath())
	if !assert.NoError(t, err) {
		t.Fatal(err)
	}
}

func TestScan(t *testing.T) {
	mainTestInit(t)
	defer mainTestCleanup(t)

	gwen := types.SERV.UserService.Get(TestUserName)
	gwenHome := gwen.GetHomeFolder()

	mediaCount := 0
	for _, child := range gwenHome.GetChildren() {

		mediaType := types.SERV.MediaRepo.TypeService().ParseExtension(
			child.Filename()[strings.Index(
				child.Filename(), ".",
			)+1:],
		)
		if mediaType.IsSupported() {
			mediaCount++
		}
	}

	task := types.SERV.TaskDispatcher.ScanDirectory(gwenHome, types.SERV.Caster)
	task.Wait()

	terr := task.ReadError()
	if !assert.Nil(t, terr) {
		t.Fatal(terr)
	}

	medias := types.SERV.MediaRepo.GetAll()
	if !assert.Equal(t, mediaCount, len(medias)) {
		t.Fatal("medias count mismatch")
	}

}
