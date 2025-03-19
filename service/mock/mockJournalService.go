package mock

import (
	"time"

	"github.com/ethanrous/weblens/fileTree"
)

type HollowJournalService struct {
	lifetimes map[string]*fileTree.Lifetime
}

func (h *HollowJournalService) GetPastFile(id fileTree.FileId, time time.Time) (*fileTree.WeblensFileImpl, error) {
	// TODO implement me
	panic("implement me")
}

func (h *HollowJournalService) Get(id fileTree.FileId) *fileTree.Lifetime {
	return h.lifetimes[id]
}

func (h *HollowJournalService) Add(lifetime ...*fileTree.Lifetime) error {
	for _, l := range lifetime {
		h.lifetimes[l.Id] = l
	}

	return nil
}

func (h *HollowJournalService) Del(id fileTree.FileId) error {
	return nil
}

func (h *HollowJournalService) IgnoreLocal() bool { return true }

func (h *HollowJournalService) SetIgnoreLocal(bool) {}

func (h *HollowJournalService) NewEvent() *fileTree.FileEvent {
	hasher := NewMockHasher()
	hasher.SetShouldCount(true)
	return fileTree.NewFileEvent(h, "", hasher)
}

func (h *HollowJournalService) WatchFolder(f *fileTree.WeblensFileImpl) error {
	return nil
}

func (h *HollowJournalService) LogEvent(fe *fileTree.FileEvent) {
	if fe == nil {
		return
	}

	for _, action := range fe.Actions {
		if action.ActionType == fileTree.FileCreate {
			lifetime, err := fileTree.NewLifetime(action)
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

func (h *HollowJournalService) GetActionsByPath(filepath fileTree.WeblensFilepath) ([]*fileTree.FileAction, error) {
	return nil, nil
}

func (h *HollowJournalService) GetPastFolderChildren(folder *fileTree.WeblensFileImpl, time time.Time) (
	[]*fileTree.WeblensFileImpl, error,
) {
	return nil, nil
}

func (h *HollowJournalService) GetLifetimesSince(date time.Time) ([]*fileTree.Lifetime, error) {
	return nil, nil
}

func (h *HollowJournalService) EventWorker() {}

func (h *HollowJournalService) FileWatcher() {}

func (h *HollowJournalService) GetActiveLifetimes() []*fileTree.Lifetime {
	return nil
}

func (h *HollowJournalService) GetAllLifetimes() []*fileTree.Lifetime {
	return nil
}

func (h *HollowJournalService) GetLifetimeByFileId(fId fileTree.FileId) *fileTree.Lifetime {
	return nil
}

func (h *HollowJournalService) GetLatestAction() (*fileTree.FileAction, error) {
	return nil, nil
}

func NewHollowJournalService() fileTree.Journal {
	return &HollowJournalService{
		lifetimes: make(map[string]*fileTree.Lifetime),
	}
}

func (h *HollowJournalService) Clear() error {
	return nil
}

func (h *HollowJournalService) UpdateLifetime(lifetime *fileTree.Lifetime) error {
	return nil
}
