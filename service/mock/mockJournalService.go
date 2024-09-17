package mock

import (
	"time"

	"github.com/ethrousseau/weblens/fileTree"
)

type HollowJournalService struct{}

func (h *HollowJournalService) GetPastFile(id fileTree.FileId, time time.Time) (*fileTree.WeblensFileImpl, error) {
	// TODO implement me
	panic("implement me")
}

func (h *HollowJournalService) Get(id fileTree.FileId) *fileTree.Lifetime {
	return nil
}

func (h *HollowJournalService) Add(lifetime *fileTree.Lifetime) error {
	return nil
}

func (h *HollowJournalService) Del(id fileTree.FileId) error {
	return nil
}

func (h *HollowJournalService) SetFileTree(ft *fileTree.FileTreeImpl) {}

func (h *HollowJournalService) NewEvent() *fileTree.FileEvent {
	return &fileTree.FileEvent{LoggedChan: make(chan struct{})}
}

func (h *HollowJournalService) WatchFolder(f *fileTree.WeblensFileImpl) error {
	return nil
}

func (h *HollowJournalService) LogEvent(fe *fileTree.FileEvent) {
	if fe != nil {
		close(fe.LoggedChan)
	}
}

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
	return &HollowJournalService{}
}
