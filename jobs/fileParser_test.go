package jobs_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ethanrous/weblens/database"
	"github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal/env"
	"github.com/ethanrous/weblens/internal/log"
	. "github.com/ethanrous/weblens/jobs"
	"github.com/ethanrous/weblens/models"
	"github.com/ethanrous/weblens/service"
	"github.com/ethanrous/weblens/service/mock"
	"github.com/ethanrous/weblens/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

var mondb *mongo.Database
var typeService models.MediaTypeService

func init() {
	var err error
	logger := log.NewZeroLogger()
	mondb, err = database.ConnectToMongo(env.GetMongoURI(), env.GetMongoDBName(env.Config{}), logger)
	if err != nil {
		panic(err)
	}

	marshMap := map[string]models.MediaType{}
	err = env.ReadTypesConfig(&marshMap)
	if err != nil {
		panic(err)
	}

	typeService = models.NewTypeService(marshMap)
}

func TestScanFile(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping %s in short mode", t.Name())
	}
	t.Parallel()

	logger := log.NewZeroLogger()

	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err != nil {
		panic(err)
	}
	defer func() { _ = col.Drop(context.Background()) }()

	testMediaTree, err := fileTree.NewFileTree(env.GetTestMediaPath(), "TEST_MEDIA", mock.NewHollowJournalService(), true, logger)
	if err != nil {
		panic(err)
	}

	if len(testMediaTree.GetRoot().GetChildren()) == 0 {
		t.Fatal("no test files found")
	}

	mediaService, err := service.NewMediaService(
		&mock.MockFileService{}, typeService, &mock.MockAlbumService{},
		col, logger,
	)
	require.NoError(t, err)

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
		err = ScanFile_(scanMeta, func() {}, logger)
		assert.NoError(t, err)
	}

	assert.Equal(t, len(testMediaTree.GetRoot().GetChildren()), mediaService.Size())
}

func TestScanDirectory(t *testing.T) {
	if testing.Short() {
		t.Skipf("skipping %s in short mode", t.Name())
	}
	t.Parallel()

	logger := log.NewZeroLogger()

	col := mondb.Collection(t.Name())
	err := col.Drop(context.Background())
	if err != nil {
		panic(err)
	}
	defer func() { _ = col.Drop(context.Background()) }()

	wp := task.NewWorkerPool(4, logger)
	wp.RegisterJob(models.ScanFileTask, ScanFile)
	wp.RegisterJob(models.ScanDirectoryTask, ScanDirectory)

	testMediaTree, err := fileTree.NewFileTree(env.GetTestMediaPath(), "TEST_MEDIA", mock.NewHollowJournalService(), true, logger)
	if err != nil {
		panic(err)
	}

	logger.Debug().Msgf("test media tree: %s", testMediaTree.GetRoot().AbsPath())

	if len(testMediaTree.GetRoot().GetChildren()) == 0 {
		t.Fatal("no test files found")
	}

	mediaService, err := service.NewMediaService(
		&mock.MockFileService{}, typeService, &mock.MockAlbumService{},
		col, logger,
	)
	require.NoError(t, err)

	wp.Run()
	defer wp.Stop()
	scanMeta := models.ScanMeta{
		File:         testMediaTree.GetRoot(),
		FileService:  &mock.MockFileService{},
		MediaService: mediaService,
		TaskService:  wp,
		Caster:       &mock.MockCaster{},
		TaskSubber:   &mock.MockClientService{},
	}

	tsk, err := wp.DispatchJob(models.ScanDirectoryTask, scanMeta, nil)
	require.NoError(t, err)

	tsk.Wait()

	_, exitStatus := tsk.Status()
	if !assert.Equal(t, task.TaskSuccess, exitStatus) {
		logger.Error().Stack().Err(tsk.ReadError()).Msg("")
		t.FailNow()
	}

	if !assert.NotNil(t, tsk.GetChildTaskPool()) {
		t.FailNow()
	}

	assert.Equal(t, 0, len(tsk.GetChildTaskPool().Errors()))
	assert.Equal(t, len(testMediaTree.GetRoot().GetChildren()), mediaService.Size())
}
