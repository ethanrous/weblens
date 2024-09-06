package proxy

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/models"
)

var _ fileTree.Journal = (*ProxyJournalService)(nil)

type ProxyJournalService struct {
	Core *models.Instance
}

func (pjs *ProxyJournalService) Get(id fileTree.FileId) *fileTree.Lifetime {
	pjs.Core.GetAddress()
	// TODO implement me
	panic("implement me")
}

func (pjs *ProxyJournalService) Add(lifetime *fileTree.Lifetime) error {
	// TODO implement me
	panic("implement me")
}

func (pjs *ProxyJournalService) Del(id fileTree.FileId) error {
	// TODO implement me
	panic("implement me")
}

func (pjs *ProxyJournalService) SetFileTree(ft *fileTree.FileTreeImpl) {
	// TODO implement me
	panic("implement me")
}

func (pjs *ProxyJournalService) NewEvent() *fileTree.FileEvent {
	// TODO implement me
	panic("implement me")
}

func (pjs *ProxyJournalService) WatchFolder(f *fileTree.WeblensFileImpl) error {
	// TODO implement me
	panic("implement me")
}

func (pjs *ProxyJournalService) LogEvent(fe *fileTree.FileEvent) {
	// TODO implement me
	panic("implement me")
}

func (pjs *ProxyJournalService) GetActionsByPath(filepath fileTree.WeblensFilepath) ([]*fileTree.FileAction, error) {
	// TODO implement me
	panic("implement me")
}

func (pjs *ProxyJournalService) GetPastFolderChildren(folder *fileTree.WeblensFileImpl, time time.Time) (
	[]*fileTree.WeblensFileImpl, error,
) {
	// TODO implement me
	panic("implement me")
}

func (pjs *ProxyJournalService) GetLatestAction() (*fileTree.FileAction, error) {
	// TODO implement me
	panic("implement me")
}

func (pjs *ProxyJournalService) GetLifetimesSince(date time.Time) ([]*fileTree.Lifetime, error) {
	endpoint := fmt.Sprintf("/history/since/%d", date.UnixMilli())
	lts, err := CallHomeStruct[[]*fileTree.Lifetime](pjs.Core, http.MethodGet, endpoint, nil)
	return lts, err
}

func (pjs *ProxyJournalService) EventWorker() {
	// TODO implement me
	panic("implement me")
}

func (pjs *ProxyJournalService) FileWatcher() {
	// TODO implement me
	panic("implement me")
}

func (pjs *ProxyJournalService) GetActiveLifetimes() []*fileTree.Lifetime {
	// TODO implement me
	panic("implement me")
}

func (pjs *ProxyJournalService) GetAllLifetimes() []*fileTree.Lifetime {
	// TODO implement me
	panic("implement me")
}
