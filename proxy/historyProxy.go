package proxy

import (
	"fmt"
	"time"

	"github.com/ethrousseau/weblens/api/fileTree"
	"github.com/ethrousseau/weblens/api/internal"
	"github.com/ethrousseau/weblens/api/types"
)

func (p *ProxyStoreImpl) GetLifetimesSince(date time.Time) ([]types.Lifetime, error) {
	timestamp := date.UnixMilli()
	if timestamp < 0 {
		timestamp = 0
	}
	
	resp, err := p.CallHome("GET", fmt.Sprintf("/api/core/history/since/%d", timestamp), nil)
	if err != nil {
		return nil, err
	}

	actions, err := ReadResponseBody[[]*fileTree.Lifetime](resp)
	if err != nil {
		return nil, err
	}

	return internal.SliceConvert[types.Lifetime](actions), nil
}

func (p *ProxyStoreImpl) WriteFileEvent(event types.FileEvent) error {
	return p.db.WriteFileEvent(event)
}

func (p *ProxyStoreImpl) GetAllLifetimes() ([]types.Lifetime, error) {
	resp, err := p.CallHome("GET", "/api/core/history", nil)
	if err != nil {
		return nil, err
	}

	lts, err := ReadResponseBody[[]*fileTree.Lifetime](resp)
	if err != nil {
		return nil, err
	}

	lifetimes := internal.SliceConvert[types.Lifetime](lts)
	// err = p.InsertManyLifetimes(lifetimes)
	// if err != nil {
	// 	return nil, err
	// }

	return lifetimes, nil
}

func (p *ProxyStoreImpl) GetLatestAction() (types.FileAction, error) {
	return p.db.GetLatestAction()
}

func (p *ProxyStoreImpl) UpsertLifetime(l types.Lifetime) error {
	return p.db.UpsertLifetime(l)
}

func (p *ProxyStoreImpl) InsertManyLifetimes(lts []types.Lifetime) error {
	return p.db.InsertManyLifetimes(lts)
}

func (p *ProxyStoreImpl) GetActionsByPath(filepath *fileTree.WeblensFilepath) ([]types.FileAction, error) {
	resp, err := p.CallHome("GET", fmt.Sprintf("/api/core/history/folder?path=%s", filepath), nil)
	if err != nil {
		return nil, err
	}

	lifetimes, err := ReadResponseBody[[]*fileTree.FileAction](resp)
	if err != nil {
		return nil, err
	}

	return internal.SliceConvert[types.FileAction](lifetimes), nil
}

func (p *ProxyStoreImpl) DeleteAllFileHistory() error {
	return p.db.DeleteAllFileHistory()
}
