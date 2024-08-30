package mock

import "github.com/ethrousseau/weblens/fileTree"

var _ fileTree.WeblensFile = (*MemFile)(nil)

type MemFile struct {
	Filename string

	buffer []byte
}

func (f *MemFile) ID() fileTree.FileId {
	return ""
}

func (f *MemFile) ReadAll() ([]byte, error) {
	return f.buffer, nil
}

func (f *MemFile) Write(data []byte) error {
	f.buffer = data
	return nil
}
