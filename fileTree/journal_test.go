package fileTree_test

import (
	"context"
	"testing"
	"time"

	"github.com/ethrousseau/weblens/database"
	. "github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal"
	"github.com/ethrousseau/weblens/internal/env"
	"github.com/ethrousseau/weblens/internal/log"
	"github.com/ethrousseau/weblens/service/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJournalImplSimple(t *testing.T) {
	mondb, err := database.ConnectToMongo(env.GetMongoURI(), env.GetMongoDBName())
	if err != nil {
		panic(err)
	}

	hasherFactory := func() Hasher {
		hasher := mock.NewMockHasher()
		hasher.SetShouldCount(true)
		return hasher
	}

	col := mondb.Collection(t.Name())
	journal, err := NewJournal(
		col,
		"weblens_test_server",
		false,
		hasherFactory,
	)
	defer journal.Close()
	defer col.Drop(context.Background())
	col.Drop(context.Background())

	tree, err := NewTestFileTree()
	require.NoError(t, err)

	tree.SetJournal(journal)
	event := journal.NewEvent()

	newDir, err := tree.MkDir(tree.GetRoot(), "newDir", event)
	require.NoError(t, err)

	journal.LogEvent(event)

	newDirLifetime := journal.Get(newDir.ID())
	var retries int
	for newDirLifetime == nil && retries < 5 {
		time.Sleep(time.Millisecond * 100)
		newDirLifetime = journal.Get(newDir.ID())
		retries++
	}
	require.NotNil(t, newDirLifetime)
	require.Equal(t, 1, len(newDirLifetime.Actions))
}

func TestJournalImpl_GetPastFile(t *testing.T) {
	mondb, err := database.ConnectToMongo(env.GetMongoURI(), env.GetMongoDBName())
	if err != nil {
		panic(err)
	}

	hasherFactory := func() Hasher {
		hasher := mock.NewMockHasher()
		hasher.SetShouldCount(true)
		return hasher
	}

	col := mondb.Collection(t.Name())
	journal, err := NewJournal(
		col,
		"weblens_test_server",
		false,
		hasherFactory,
	)
	defer journal.Close()
	defer col.Drop(context.Background())
	col.Drop(context.Background())

	tree, err := NewTestFileTree()
	require.NoError(t, err)

	tree.SetJournal(journal)
	event := journal.NewEvent()

	event.NewCreateAction(tree.GetRoot())

	newDir, err := tree.MkDir(tree.GetRoot(), "newDir", event)
	require.NoError(t, err)

	testFile, err := tree.Touch(newDir, "test_file", event)
	require.NoError(t, err)

	log.Trace.Println("Logging event")
	journal.LogEvent(event)
	event.Wait()

	deleteEvent := journal.NewEvent()
	tree.Delete(testFile.ID(), deleteEvent)
	tree.Delete(newDir.ID(), deleteEvent)

	journal.LogEvent(deleteEvent)

	assert.Empty(t, tree.GetRoot().GetChildren())

	// Get the root directory just before the delete event
	pastRoot, err := journal.GetPastFile(tree.GetRoot().ID(), deleteEvent.EventBegin.Add(-time.Microsecond))
	require.NoError(t, err)
	assert.Equal(t, tree.GetRoot().ID(), pastRoot.ID())

	pastRootChildren, err := journal.GetPastFolderChildren(
		tree.GetRoot(), deleteEvent.EventBegin.Add(-time.Microsecond),
	)
	require.NoError(t, err)

	if !assert.Equal(t, 1, len(pastRootChildren)) {
		childNames := internal.Map(
			pastRootChildren, func(f *WeblensFileImpl) string {
				return f.Name()
			},
		)
		log.Debug.Println("Wrong children:", childNames)
		t.FailNow()
	}
	assert.Equal(t, newDir.ID(), pastRootChildren[0].ID())

	// Get the old folder just before the delete event
	pastDir, err := journal.GetPastFile(newDir.ID(), deleteEvent.EventBegin.Add(-time.Microsecond))
	require.NoError(t, err)

	require.Equal(t, newDir.ID(), pastDir.ID())

	pastDirChildren, err := journal.GetPastFolderChildren(newDir, deleteEvent.EventBegin.Add(-time.Microsecond))
	require.NoError(t, err)

	require.Equal(t, 1, len(pastDirChildren))
	assert.Equal(t, testFile.ID(), pastDirChildren[0].ID())
}
