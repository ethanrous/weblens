package history

import (
	"slices"

	"github.com/ethrousseau/weblens/api/types"
)

type journalService struct {
	lifetimes    map[types.LifetimeId]types.Lifetime
	latestUpdate map[types.FileId]types.Lifetime
}

func NewService(fileTree types.FileTree, dbServer types.DatabaseService) types.JournalService {
	if fileTree == nil || dbServer == nil {
		return nil
	}
	return &journalService{
		lifetimes:    make(map[types.LifetimeId]types.Lifetime),
		latestUpdate: make(map[types.FileId]types.Lifetime),
	}
}

func (j *journalService) Init(db types.DatabaseService) error {
	lifetimes, err := db.GetAllLifetimes()
	if err != nil {
		return err
	}

	for _, lt := range lifetimes {
		j.lifetimes[lt.ID()] = lt
		if lt.GetLatestFileId() != "" {
			j.latestUpdate[lt.GetLatestFileId()] = lt
		}
	}
	// util.Debug.Println(events)

	return nil
}

func (j *journalService) GetActiveLifetimes() []types.Lifetime {
	var result []types.Lifetime
	for _, l := range j.lifetimes {
		if l.IsLive() {
			result = append(result, l)
		}
	}
	return result
}

func (j *journalService) GetAllFileEvents() ([]types.FileEvent, error) {
	// return util.MapToSlicePure(j.)
	// return types.SERV.Database.GetAllFileEvents()
	return nil, types.ErrNotImplemented("get all file events")
}

func (j *journalService) LogEvent(fe types.FileEvent) error {
	if len(fe.GetActions()) == 0 {
		return nil
	}

	actions := fe.GetActions()
	slices.SortFunc(
		actions, func(a, b types.FileAction) int {
			return a.GetTimestamp().Compare(b.GetTimestamp())
		},
	)

	var updated []types.Lifetime

	for _, action := range fe.GetActions() {
		switch action.GetActionType() {
		case FileCreate:
			newL, err := NewLifetime("", action)
			if err != nil {
				return err
			}

			if f := types.SERV.FileTree.Get(action.GetDestinationId()); !f.IsDir() {
				contentId := f.GetContentId()
				if contentId == "" {
					return types.NewWeblensError("No content ID in lifetime update")
				}
				newL.SetContentId(contentId)
			}

			j.lifetimes[newL.ID()] = newL
			j.latestUpdate[action.GetDestinationId()] = newL
			updated = append(updated, newL)
		case FileMove:
			existing := j.latestUpdate[action.GetOriginId()]
			existing.Add(action)
			updated = append(updated, existing)
		}
	}

	for _, update := range updated {
		if update.GetContentId() == "" && !types.SERV.FileTree.Get(update.GetLatestFileId()).IsDir() {
			return types.NewWeblensError("No content ID in lifetime update")
		}
		err := types.SERV.Database.AddOrUpdateLifetime(update)
		if err != nil {
			return err
		}
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
