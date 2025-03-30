package journal

import (
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/models/history"
	"github.com/ethanrous/weblens/modules/fs"
)

type HollowJournalService struct {
	lifetimes map[string]*history.Lifetime
}

func (h *HollowJournalService) GetPastFile(id string, time time.Time) (*file_model.WeblensFileImpl, error) {
	// TODO implement me
	panic("implement me")
}

func (h *HollowJournalService) Get(id string) *history.Lifetime {
	return h.lifetimes[id]
}

func (h *HollowJournalService) Add(lifetime ...*history.Lifetime) error {
	for _, l := range lifetime {
		h.lifetimes[l.Id] = l
	}

	return nil
}

func (h *HollowJournalService) Del(id string) error {
	return nil
}

func (h *HollowJournalService) IgnoreLocal() bool { return true }

func (h *HollowJournalService) SetIgnoreLocal(bool) {}

func (h *HollowJournalService) NewEvent() *history.FileEvent {
	hasher := NewMockHasher()
	hasher.SetShouldCount(true)
	return history.NewFileEvent(h, "", hasher)
}

func (h *HollowJournalService) WatchFolder(f *file_model.WeblensFileImpl) error {
	return nil
}

func (h *HollowJournalService) LogEvent(fe *history.FileEvent) {
	if fe == nil {
		return
	}

	for _, action := range fe.Actions {
		if action.ActionType == history.FileCreate {
			lifetime, err := history.NewLifetime(action)
			if err != nil {
				continue
			}
			h.lifetimes[lifetime.ID()] = lifetime
		} else {
			lifetime, ok := h.lifetimes[action.LifeId]
			if !ok {
				continue
			}
			lifetime.Actions = append(lifetime.Actions, action)
		}
	}

	close(fe.LoggedChan)
}

func (h *HollowJournalService) Flush() {}

func (h *HollowJournalService) GetActionsByPath(filepath fs.Filepath) ([]*history.FileAction, error) {
	return nil, nil
}

func (h *HollowJournalService) GetPastFolderChildren(folder *file_model.WeblensFileImpl, time time.Time) (
	[]*file_model.WeblensFileImpl, error,
) {
	return nil, nil
}

func (h *HollowJournalService) GetLifetimesSince(date time.Time) ([]*history.Lifetime, error) {
	return nil, nil
}

func (h *HollowJournalService) EventWorker() {}

func (h *HollowJournalService) FileWatcher() {}

func (h *HollowJournalService) GetActiveLifetimes() []*history.Lifetime {
	return nil
}

func (h *HollowJournalService) GetAllLifetimes() []*history.Lifetime {
	return nil
}

func (h *HollowJournalService) GetLifetimeByFileId(fId string) *history.Lifetime {
	return nil
}

func (h *HollowJournalService) GetLatestAction() (*history.FileAction, error) {
	return nil, nil
}

func NewHollowJournalService() history.Journal {
	return &HollowJournalService{
		lifetimes: make(map[string]*history.Lifetime),
	}
}

func (h *HollowJournalService) Clear() error {
	return nil
}

func (h *HollowJournalService) UpdateLifetime(lifetime *history.Lifetime) error {
	return nil
}
