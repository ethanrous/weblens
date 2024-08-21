package proxy

import (
	"fmt"
	"io"
	"os"

	"github.com/ethrousseau/weblens/api/fileTree"
	"github.com/ethrousseau/weblens/api/internal"
	error2 "github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/types"
)

func (p *ProxyStoreImpl) NewTrashEntry(te types.TrashEntry) error {
	return error2.NotImplemented("NewTrashEntry")
}

func (p *ProxyStoreImpl) DeleteTrashEntry(fileId types.FileId) error {
	return error2.NotImplemented("DeleteTrashEntry")
}

func (p *ProxyStoreImpl) GetTrashEntry(fileId types.FileId) (te types.TrashEntry, err error) {
	return te, error2.NotImplemented("GetTrashEntry")
}

func (p *ProxyStoreImpl) GetAllFiles() ([]*fileTree.WeblensFile, error) {
	resp, err := p.CallHome("GET", "/api/core/files", nil)
	if err != nil {
		return nil, err
	}
	files, err := ReadResponseBody[[]*fileTree.WeblensFile](resp)
	if err != nil {
		return nil, err
	}

	return internal.SliceConvert[*fileTree.WeblensFile](files), nil
}

func (p *ProxyStoreImpl) StatFile(f *fileTree.WeblensFile) (types.FileStat, error) {
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

func (p *ProxyStoreImpl) ReadFile(f *fileTree.WeblensFile) ([]byte, error) {
	resp, err := p.CallHome("GET", fmt.Sprintf("/api/core/file/%s/content", f.ID()), nil)
	if err != nil {
		return nil, err
	}

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, error2.Wrap(err)
	}

	return bs, nil
}

func (p *ProxyStoreImpl) StreamFile(f *fileTree.WeblensFile) (io.ReadCloser, error) {
	resp, err := p.CallHome("GET", fmt.Sprintf("/api/core/file/%s/content", f.ID()), nil)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (p *ProxyStoreImpl) ReadDir(f *fileTree.WeblensFile) ([]types.FileStat, error) {
	if !f.IsDir() {
		return nil, error2.WErrMsg("trying to read directory on regular file")
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

func (p *ProxyStoreImpl) TouchFile(f *fileTree.WeblensFile) error {
	stat, _ := p.db.StatFile(f)
	if stat.Exists {
		return types.ErrFileAlreadyExists(f.GetAbsPath())
	}

	var err error
	if f.IsDir() {
		err = os.Mkdir(f.GetAbsPath(), os.FileMode(0777))
		if err != nil {
			return error2.Wrap(err)
		}
	} else {
		var osFile *os.File
		osFile, err = os.Create(f.GetAbsPath())
		err = osFile.Close()
		if err != nil {
			return error2.Wrap(err)
		}
	}

	return nil
}

func (p *ProxyStoreImpl) GetFile(fileId types.FileId) (*fileTree.WeblensFile, error) {
	resp, err := p.CallHome("GET", fmt.Sprintf("/api/core/file/%s", fileId), nil)
	if err != nil {
		return nil, err
	}

	file, err := ReadResponseBody[*fileTree.WeblensFile](resp)
	if err != nil {
		return nil, err
	}

	return file, nil
}
