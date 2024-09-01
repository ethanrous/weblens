package jobs_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ethrousseau/weblens/database"
	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	. "github.com/ethrousseau/weblens/jobs"
	"github.com/ethrousseau/weblens/models"
	"github.com/ethrousseau/weblens/service"
	"github.com/ethrousseau/weblens/service/mock"
	"github.com/ethrousseau/weblens/task"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

var mondb *mongo.Database
var typeService models.MediaTypeService

func init() {
	log.SetLogLevel(log.DEBUG)

	var err error
	mondb, err = database.ConnectToMongo(internal.GetMongoURI(), internal.GetMongoDBName()+"-test")
	if err != nil {
		panic(err)
	}

	marshMap := map[string]models.MediaType{}
	internal.ReadTypesConfig(&marshMap)
	typeService = models.NewTypeService(marshMap)
}

func TestScanFile(t *testing.T) {
	t.Parallel()

	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err != nil {
		panic(err)
	}
	defer col.Drop(context.Background())

	testMediaTree, err := fileTree.NewFileTree(
		internal.GetTestMediaPath(), "TEST_MEDIA", mock.NewMockHasher(), mock.NewHollowJournalService(),
	)
	if err != nil {
		panic(err)
	}

	mediaService, err := service.NewMediaService(
		&mock.MockFileService{}, typeService, &mock.MockAlbumService{},
		col,
	)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	for _, file := range testMediaTree.GetRoot().GetChildren() {
		ext := filepath.Ext(file.Filename())
		if !mediaService.GetMediaTypes().ParseExtension(ext).Displayable {
			continue
		}
		scanMeta := models.ScanMeta{
			File:         file,
			FileService:  &mock.MockFileService{},
			MediaService: mediaService,
			Caster:       &mock.MockCaster{},
		}
		err = ScanFile_(scanMeta)
		assert.NoError(t, err)
	}

	assert.Equal(t, len(testMediaTree.GetRoot().GetChildren()), mediaService.Size())
}

func TestScanDirectory(t *testing.T) {
	t.Parallel()

	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err != nil {
		panic(err)
	}
	defer col.Drop(context.Background())

	wp := task.NewWorkerPool(4, -1)
	wp.RegisterJob(models.ScanFileTask, ScanFile)
	wp.RegisterJob(models.ScanDirectoryTask, ScanDirectory)

	testMediaTree, err := fileTree.NewFileTree(
		internal.GetTestMediaPath(), "TEST_MEDIA", mock.NewMockHasher(), mock.NewHollowJournalService(),
	)
	if err != nil {
		panic(err)
	}

	mediaService, err := service.NewMediaService(
		&mock.MockFileService{}, typeService, &mock.MockAlbumService{},
		col,
	)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	wp.Run()
	scanMeta := models.ScanMeta{
		File:         testMediaTree.GetRoot(),
		FileService:  &mock.MockFileService{},
		MediaService: mediaService,
		TaskService:  wp,
		Caster:       &mock.MockCaster{},
		TaskSubber:   &mock.MockClientService{},
	}

	tsk, err := wp.DispatchJob(models.ScanDirectoryTask, scanMeta, nil)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	tsk.Wait()

	_, exitStatus := tsk.Status()
	if !assert.Equal(t, task.TaskSuccess, exitStatus) {
		t.FailNow()
	}

	if !assert.NotNil(t, tsk.GetChildTaskPool()) {
		t.FailNow()
	}

	assert.Equal(t, 0, len(tsk.GetChildTaskPool().Errors()))
	assert.Equal(t, len(testMediaTree.GetRoot().GetChildren()), mediaService.Size())
}
