package jobs_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ethrousseau/weblens/database"
	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/jobs"
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
	if internal.IsDevMode() {
		log.DoDebug()
	}

	mondb = database.ConnectToMongo("mongodb://localhost:27017", "weblens-test")

	marshMap := map[string]models.MediaType{}
	internal.ReadTypesConfig(&marshMap)
	typeService = models.NewTypeService(marshMap)
}

func TestFileScan(t *testing.T) {
	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err != nil {
		panic(err)
	}
	defer col.Drop(context.Background())

	wp := task.NewWorkerPool(4)
	wp.RegisterJob(models.ScanFileTask, jobs.ScanFile)

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

	pool := wp.NewTaskPool(false, nil)
	wp.Run()

	var queuedCount int
	for _, file := range testMediaTree.GetRoot().GetChildren() {
		ext := filepath.Ext(file.Filename())
		if !mediaService.GetMediaTypes().ParseExtension(ext).Displayable {
			continue
		}
		subMeta := models.ScanMeta{
			File:         file,
			FileService:  &mock.MockFileService{},
			MediaService: mediaService,
			TaskService:  wp,
			Caster:       &mock.MockCaster{},
		}

		_, err = wp.DispatchJob(models.ScanFileTask, subMeta, pool)
		assert.NoError(t, err)
		queuedCount++
	}
	pool.SignalAllQueued()
	pool.Wait(false)
	assert.Equal(t, 0, len(pool.Errors()))
	assert.Equal(t, queuedCount, mediaService.Size())
}

func TestScanDirectory(t *testing.T) {
	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err != nil {
		panic(err)
	}
	defer col.Drop(context.Background())

	wp := task.NewWorkerPool(4)
	wp.RegisterJob(models.ScanFileTask, jobs.ScanFile)
	wp.RegisterJob(models.ScanDirectoryTask, jobs.ScanDirectory)

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
	subMeta := models.ScanMeta{
		File:         testMediaTree.GetRoot(),
		FileService:  &mock.MockFileService{},
		MediaService: mediaService,
		TaskService:  wp,
		Caster:       &mock.MockCaster{},
		TaskSubber:   &mock.MockClientService{},
	}

	tsk, err := wp.DispatchJob(models.ScanDirectoryTask, subMeta, nil)
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
