package proxy

import (
	"fmt"
	"time"

	"github.com/ethrousseau/weblens/api/dataStore/history"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

func (p *ProxyStore) GetLifetimesSince(date time.Time) ([]types.Lifetime, error) {
	timestamp := date.UnixMilli()
	if timestamp < 0 {
		timestamp = 0
	}
	
	resp, err := p.CallHome("GET", fmt.Sprintf("/api/core/history/since/%d", timestamp), nil)
	if err != nil {
		return nil, err
	}

	actions, err := ReadResponseBody[[]*history.Lifetime](resp)
	if err != nil {
		return nil, err
	}

	return util.SliceConvert[types.Lifetime](actions), nil
}

func (p *ProxyStore) WriteFileEvent(event types.FileEvent) error {
	return p.db.WriteFileEvent(event)
}

func (p *ProxyStore) GetAllLifetimes() ([]types.Lifetime, error) {
	resp, err := p.CallHome("GET", "/api/core/history", nil)
	if err != nil {
		return nil, err
	}

	lts, err := ReadResponseBody[[]*history.Lifetime](resp)
	if err != nil {
		return nil, err
	}

	lifetimes := util.SliceConvert[types.Lifetime](lts)
	// err = p.InsertManyLifetimes(lifetimes)
	// if err != nil {
	// 	return nil, err
	// }

	return lifetimes, nil
}

func (p *ProxyStore) GetLatestAction() (types.FileAction, error) {
	return p.db.GetLatestAction()
}

func (p *ProxyStore) UpsertLifetime(l types.Lifetime) error {
	return p.db.UpsertLifetime(l)
}

func (p *ProxyStore) InsertManyLifetimes(lts []types.Lifetime) error {
	return p.db.InsertManyLifetimes(lts)
}

func (p *ProxyStore) GetActionsByPath(filepath types.WeblensFilepath) ([]types.FileAction, error) {
	resp, err := p.CallHome("GET", fmt.Sprintf("/api/core/history/folder?path=%s", filepath), nil)
	if err != nil {
		return nil, err
	}

	lifetimes, err := ReadResponseBody[[]*history.FileAction](resp)
	if err != nil {
		return nil, err
	}

	return util.SliceConvert[types.FileAction](lifetimes), nil
}

func (p *ProxyStore) DeleteAllFileHistory() error {
	return p.db.DeleteAllFileHistory()
}
