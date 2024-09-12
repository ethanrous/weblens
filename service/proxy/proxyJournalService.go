package proxy

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ethrousseau/weblens/fileTree"
	"github.com/ethrousseau/weblens/internal/werror"
	"github.com/ethrousseau/weblens/models"
)

var _ fileTree.Journal = (*ProxyJournalService)(nil)

type ProxyJournalService struct {
	Core *models.Instance
}

func (pjs *ProxyJournalService) Get(id fileTree.FileId) *fileTree.Lifetime {
	pjs.Core.GetAddress()
	
	panic("implement me")
}

func (pjs *ProxyJournalService) Add(lifetime *fileTree.Lifetime) error {
	
	panic("implement me")
}

func (pjs *ProxyJournalService) Del(id fileTree.FileId) error {
	
	panic("implement me")
}

func (pjs *ProxyJournalService) SetFileTree(ft *fileTree.FileTreeImpl) {
	
	panic("implement me")
}

func (pjs *ProxyJournalService) NewEvent() *fileTree.FileEvent {
	
	panic("implement me")
}

func (pjs *ProxyJournalService) WatchFolder(f *fileTree.WeblensFileImpl) error {
	
	panic("implement me")
}

func (pjs *ProxyJournalService) LogEvent(fe *fileTree.FileEvent) {
	
	panic("implement me")
}

func (pjs *ProxyJournalService) GetActionsByPath(filepath fileTree.WeblensFilepath) ([]*fileTree.FileAction, error) {
	
	panic("implement me")
}

func (pjs *ProxyJournalService) GetPastFolderChildren(folder *fileTree.WeblensFileImpl, time time.Time) (
	[]*fileTree.WeblensFileImpl, error,
) {
	
	panic("implement me")
}

func (pjs *ProxyJournalService) GetLatestAction() (*fileTree.FileAction, error) {
	
	panic("implement me")
}

func (pjs *ProxyJournalService) GetLifetimesSince(date time.Time) ([]*fileTree.Lifetime, error) {
	millis := date.UnixMilli()
	if millis < 0 {
		return nil, werror.Errorf("Trying to get lifetimes with millis less than 0")
	}

	endpoint := fmt.Sprintf("/history/since/%d", millis)
	lts, err := CallHomeStruct[[]*fileTree.Lifetime](pjs.Core, http.MethodGet, endpoint, nil)
	return lts, err
}

func (pjs *ProxyJournalService) EventWorker() {
	
	panic("implement me")
}

func (pjs *ProxyJournalService) FileWatcher() {
	
	panic("implement me")
}

func (pjs *ProxyJournalService) GetActiveLifetimes() []*fileTree.Lifetime {
	
	panic("implement me")
}

func (pjs *ProxyJournalService) GetAllLifetimes() []*fileTree.Lifetime {
	
	panic("implement me")
}
