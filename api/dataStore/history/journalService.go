package history

import (
	"slices"

	"github.com/ethrousseau/weblens/api/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type journalService struct {
	lifetimes    map[types.LifetimeId][]types.FileAction
	latestUpdate map[types.FileId]types.LifetimeId

	fileTree types.FileTree
	dbServer types.DatabaseService
}

type lifetime []types.FileAction

func NewJournalService(fileTree types.FileTree, dbServer types.DatabaseService) types.JournalService {
	if fileTree == nil || dbServer == nil {
		return nil
	}
	return &journalService{fileTree: fileTree, dbServer: dbServer}
}

func (j *journalService) LogEvent(fe types.FileEvent) error {
	if len(fe.GetActions()) == 0 {
		return nil
	}

	actions := fe.GetActions()
	slices.SortFunc(actions, func(a, b types.FileAction) int {
		return a.GetTimestamp().Compare(b.GetTimestamp())
	})

	var updated []types.LifetimeId

	for _, action := range fe.GetActions() {
		switch action.GetActionType() {
		case FileCreate:
			newLId := types.LifetimeId(primitive.NewObjectID().String())
			j.lifetimes[newLId] = []types.FileAction{action}
			j.latestUpdate[action.GetDestinationId()] = newLId
			updated = append(updated, newLId)
		case FileMove:
			lId := j.latestUpdate[action.GetOriginId()]
			l := j.lifetimes[lId]
			l = append(l, action)
			j.lifetimes[lId] = l
			updated = append(updated, lId)
		}
	}

	err := j.dbServer.WriteFileEvent(fe)
	if err != nil {
		return err
	}

	return nil
}
