package history

import (
	"slices"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type journalService struct {
	lifetimes    map[types.LifetimeId][]types.FileAction
	latestUpdate map[types.FileId]types.LifetimeId
}

func NewService(fileTree types.FileTree, dbServer types.DatabaseService) types.JournalService {
	if fileTree == nil || dbServer == nil {
		return nil
	}
	return &journalService{
		lifetimes:    make(map[types.LifetimeId][]types.FileAction),
		latestUpdate: make(map[types.FileId]types.LifetimeId),
	}
}

func (j *journalService) Init(db types.DatabaseService) error {
	events, err := db.GetAllLifetimes()
	if err != nil {
		return err
	}

	util.Debug.Println(events)

	return nil
}

func (j *journalService) GetActiveLifetimes() []types.Lifetime {
	var result []types.Lifetime
	for _, l := range j.lifetimes {
		if l[len(l)-1].GetActionType() != FileDelete {
			result = append(result, lifetime{
				fileId:    l[len(l)-1].GetDestinationId(),
				contentId: l[len(l)-1].GetContentId(),
			})
		}
	}
	return result
}

func (j *journalService) GetAllFileEvents() ([]types.FileEvent, error) {
	// return util.MapToSlicePure(j.)
	// return types.SERV.Database.GetAllFileEvents()
	return nil, nil
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

	err := types.SERV.Database.WriteFileEvent(fe)
	if err != nil {
		return err
	}

	return nil
}

func (j *journalService) Get(lId types.LifetimeId) []types.FileAction {
	panic(types.ErrNotImplemented("journal get"))
	return nil
}

func (j *journalService) Add([]types.FileAction) error {
	return types.ErrNotImplemented("journal add")
}

func (j *journalService) Del(lId types.LifetimeId) error {
	return types.ErrNotImplemented("journal del")
}

func (j *journalService) Size() int {
	return len(j.lifetimes)
}
