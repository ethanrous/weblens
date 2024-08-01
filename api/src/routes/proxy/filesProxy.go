package proxy

import (
	"fmt"
	"io"

	"github.com/ethrousseau/weblens/api/dataStore/filetree"
	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
)

func (p *ProxyStore) NewTrashEntry(te types.TrashEntry) error {
	return types.ErrNotImplemented("NewTrashEntry")
}

func (p *ProxyStore) DeleteTrashEntry(fileId types.FileId) error {
	return types.ErrNotImplemented("DeleteTrashEntry")
}

func (p *ProxyStore) GetTrashEntry(fileId types.FileId) (te types.TrashEntry, err error) {
	return te, types.ErrNotImplemented("GetTrashEntry")
}

func (p *ProxyStore) GetAllFiles() ([]types.WeblensFile, error) {
	resp, err := p.CallHome("GET", "/api/core/files", nil)
	if err != nil {
		return nil, err
	}
	files, err := ReadResponseBody[[]*filetree.WeblensFile](resp)
	if err != nil {
		return nil, err
	}

	return util.SliceConvert[types.WeblensFile](files), nil
}

func (p *ProxyStore) StatFile(f types.WeblensFile) (types.FileStat, error) {
	resp, err := p.CallHome("GET", fmt.Sprintf("/api/core/file/%s/stat", f.ID()), nil)
	if err != nil {
		return types.FileStat{}, err
	}

	stat, err := ReadResponseBody[types.FileStat](resp)
	if err != nil {
		return types.FileStat{}, err
	}

	return stat, nil
}

func (p *ProxyStore) ReadFile(f types.WeblensFile) ([]byte, error) {
	resp, err := p.CallHome("GET", fmt.Sprintf("/api/core/file/%s/content", f.ID()), nil)
	if err != nil {
		return nil, err
	}

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, types.WeblensErrorFromError(err)
	}

	return bs, nil
}

func (p *ProxyStore) ReadDir(f types.WeblensFile) ([]types.FileStat, error) {
	if !f.IsDir() {
		return nil, types.WeblensErrorMsg("trying to read directory on regular file")
	}

	resp, err := p.CallHome("GET", fmt.Sprintf("/api/core/file/%s/directory", f.ID()), nil)
	if err != nil {
		return nil, err
	}

	children, err := ReadResponseBody[[]types.FileStat](resp)
	if err != nil {
		return nil, err
	}

	return children, nil
}

func (p *ProxyStore) TouchFile(f types.WeblensFile) error {
	return types.ErrNotImplemented("TouchFile")
}

func (p *ProxyStore) GetFile(fileId types.FileId) (types.WeblensFile, error) {
	resp, err := p.CallHome("GET", fmt.Sprintf("/api/core/file/%s", fileId), nil)
	if err != nil {
		return nil, err
	}

	file, err := ReadResponseBody[*filetree.WeblensFile](resp)
	if err != nil {
		return nil, err
	}

	return file, nil
}
