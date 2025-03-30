package history

import (
	"context"
	"testing"
	"time"

	"github.com/ethanrous/weblens/database"
	. "github.com/ethanrous/weblens/fileTree"
	"github.com/ethanrous/weblens/internal"
	"github.com/ethanrous/weblens/internal/env"
	"github.com/ethanrous/weblens/internal/log"
	"github.com/ethanrous/weblens/service/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJournalImplSimple(t *testing.T) {
	logger := log.NewZeroLogger()
	mondb, err := database.ConnectToMongo(env.GetMongoURI(), env.GetMongoDBName(env.Config{}), logger)
	if err != nil {
		panic(err)
	}

	hasherFactory := func() Hasher {
		hasher := mock.NewMockHasher()
		hasher.SetShouldCount(true)
		return hasher
	}

	col := mondb.Collection(t.Name())
	journalConfig := JournalConfig{
		Collection:    col,
		ServerId:      "weblens_test_server",
		IgnoreLocal:   false,
		HasherFactory: hasherFactory,
		Logger:        logger,
	}
	journal, err := NewJournal(journalConfig)
	require.NoError(t, err)
	defer journal.Close()
	err = col.Drop(context.Background())
	require.NoError(t, err)
	defer col.Drop(context.Background())

	tree, err := NewTestFileTree()
	require.NoError(t, err)

	tree.SetJournal(journal)
	event := journal.NewEvent()

	newDir, err := tree.MkDir(tree.GetRoot(), "newDir", event)
	require.NoError(t, err)

	journal.LogEvent(event)
	event.Wait()

	newDirLifetime := journal.Get(newDir.ID())
	// var retries int
	// for newDirLifetime == nil && retries < 5 {
	// 	time.Sleep(time.Millisecond * 20)
	// 	newDirLifetime = journal.Get(newDir.ID())
	// 	retries++
	// }
	require.NotNil(t, newDirLifetime)
	require.Equal(t, 1, len(newDirLifetime.Actions))
}

func TestJournalImpl_GetPastFile(t *testing.T) {
	logger := log.NewZeroLogger()

	mondb, err := database.ConnectToMongo(env.GetMongoURI(), env.GetMongoDBName(env.Config{}), logger)
	if err != nil {
		panic(err)
	}

	hasherFactory := func() Hasher {
		hasher := mock.NewMockHasher()
		hasher.SetShouldCount(true)
		return hasher
	}

	col := mondb.Collection(t.Name() + "-journal")
	journalConfig := JournalConfig{
		Collection:    col,
		ServerId:      "weblens_test_server",
		IgnoreLocal:   false,
		HasherFactory: hasherFactory,
		Logger:        logger,
	}
	journal, err := NewJournal(journalConfig)
	require.NoError(t, err)

	defer journal.Close()
	err = col.Drop(context.Background())
	require.NoError(t, err)
	defer col.Drop(context.Background())

	tree, err := NewTestFileTree()
	require.NoError(t, err)

	tree.SetJournal(journal)
	event := journal.NewEvent()

	event.NewCreateAction(tree.GetRoot())

	newDir, err := tree.MkDir(tree.GetRoot(), "newDir", event)
	require.NoError(t, err)

	testFile, err := tree.Touch(newDir, "test_file", event)
	require.NoError(t, err)

	journal.LogEvent(event)
	event.Wait()

	// Mongo only has millisecond precision, so we need to wait a bit to ensure the delete event is
	// for sure after the create events. This test is a bit flaky without this. In reality, this is
	// not an issue because two related events will never be within the same millisecond as they are here.
	time.Sleep(time.Millisecond * 10)

	secondDirEvent := journal.NewEvent()
	newDir2, err := tree.MkDir(tree.GetRoot(), "newDir2", secondDirEvent)
	require.NoError(t, err)

	journal.LogEvent(secondDirEvent)
	secondDirEvent.Wait()

	pastRootChildren, err := journal.GetPastFolderChildren(
		tree.GetRoot(), secondDirEvent.EventBegin.Add(-time.Millisecond*2),
	)
	require.NoError(t, err)

	if !assert.Equal(t, 1, len(pastRootChildren)) {
		childNames := internal.Map(
			pastRootChildren, func(f *WeblensFileImpl) string {
				return f.Name()
			},
		)
		logger.Error().Msgf("Wrong children: %v", childNames)
		t.FailNow()
	}

	time.Sleep(time.Millisecond * 10)

	deleteEvent := journal.NewEvent()
	err = tree.Delete(testFile.ID(), deleteEvent)
	require.NoError(t, err)
	err = tree.Delete(newDir.ID(), deleteEvent)
	require.NoError(t, err)
	err = tree.Delete(newDir2.ID(), deleteEvent)
	require.NoError(t, err)

	journal.LogEvent(deleteEvent)
	deleteEvent.Wait()

	assert.Empty(t, tree.GetRoot().GetChildren())

	// Get the root directory just before the delete event
	pastRoot, err := journal.GetPastFile(tree.GetRoot().ID(), deleteEvent.EventBegin)
	require.NoError(t, err)
	assert.Equal(t, tree.GetRoot().ID(), pastRoot.ID())

	pastRootChildren, err = journal.GetPastFolderChildren(
		tree.GetRoot(), deleteEvent.EventBegin.Add(-time.Millisecond),
	)
	require.NoError(t, err)

	if !assert.Equal(t, 2, len(pastRootChildren)) {
		for _, lt := range journal.GetAllLifetimes() {
			log.Error.Printf("Lifetime [%s]: [%s]", lt.Actions[0].DestinationPath, lt.Actions[0].Timestamp)
		}

		log.Error.Println("Delete Start:", deleteEvent.EventBegin)
		log.Error.Println("Second Dir Start:", secondDirEvent.EventBegin)
		t.FailNow()
	}
	assert.Contains(t, []string{newDir.ID(), newDir2.ID()}, pastRootChildren[0].ID())
	assert.Contains(t, []string{newDir.ID(), newDir2.ID()}, pastRootChildren[1].ID())

	getChildrenAfterDeleteTime := time.Now()
	// Get the root children at the current time
	currentChildren, err := journal.GetPastFolderChildren(
		tree.GetRoot(), getChildrenAfterDeleteTime,
	)
	require.NoError(t, err)

	// Make sure the journal thinks the current root has no children
	if !assert.Empty(t, currentChildren) {
		log.Error.Printf("Current Children at Delete Time: %s", getChildrenAfterDeleteTime)
		for _, child := range currentChildren {
			log.Error.Printf("Current Child: %s", child.Name())
		}
		t.FailNow()
	}

	// Get the old folder just before the delete event
	pastDir, err := journal.GetPastFile(newDir.ID(), deleteEvent.EventBegin)
	require.NoError(t, err)

	require.Equal(t, newDir.ID(), pastDir.ID())

	pastDirChildren, err := journal.GetPastFolderChildren(newDir, deleteEvent.EventBegin.Add(-time.Millisecond))
	require.NoError(t, err)

	require.Equal(t, 1, len(pastDirChildren))
	assert.Equal(t, testFile.ID(), pastDirChildren[0].ID())
}
