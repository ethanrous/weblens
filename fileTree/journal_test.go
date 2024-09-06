package fileTree_test

import (
	"testing"
	"time"

	"github.com/ethrousseau/weblens/database"
	. "github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal/env"
	"github.com/stretchr/testify/require"
)

func TestJournalImplSimple(t *testing.T) {
	mondb, err := database.ConnectToMongo(env.GetMongoURI(), env.GetMongoDBName())
	if err != nil {
		panic(err)
	}

	journal, err := NewJournal(
		mondb.Collection(t.Name()),
		"weblens_test_server",
	)
	defer journal.Close()

	tree, err := newTestFileTree()
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
