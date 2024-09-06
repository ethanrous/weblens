package mock

import (
	"strconv"
	"sync"

	"github.com/ethrousseau/weblens/fileTree"
)

type MockHasher struct {
	mu    sync.Mutex
	count int64
}

func NewMockHasher() fileTree.Hasher {
	return &MockHasher{}
}

func (h *MockHasher) Hash(file *fileTree.WeblensFileImpl, event *fileTree.FileEvent) error {
	h.mu.Lock()
	file.SetContentId(strconv.FormatInt(h.count, 10))
	h.count++
	h.mu.Unlock()
	event.NewCreateAction(file)

	return nil
}
