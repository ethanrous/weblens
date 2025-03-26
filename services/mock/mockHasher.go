package mock

import (
	"strconv"
	"sync"

	"github.com/ethanrous/weblens/fileTree"
)

type MockHasher struct {
	mu          sync.Mutex
	count       int64
	shouldCount bool
}

func NewMockHasher() *MockHasher {
	return &MockHasher{}
}

func (h *MockHasher) Hash(file *fileTree.WeblensFileImpl) error {
	if h.shouldCount {
		h.mu.Lock()
		file.SetContentId(strconv.FormatInt(h.count, 10))
		h.count++
		h.mu.Unlock()
	}

	return nil
}

func (h *MockHasher) SetShouldCount(shouldCount bool) {
	h.shouldCount = shouldCount
}
