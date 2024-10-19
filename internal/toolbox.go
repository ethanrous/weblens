package internal

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ethanrous/weblens/internal/log"
)

type WeblensHash struct {
	hash hash.Hash
}

func NewWeblensHash() *WeblensHash {
	return &WeblensHash{hash: sha256.New()}
}

func (h *WeblensHash) Add(data []byte) error {
	_, err := h.hash.Write(data)
	return err
}

func (h *WeblensHash) Done(len int) string {
	return base64.URLEncoding.EncodeToString(h.hash.Sum(nil))[:len]
}

// GlobbyHash Set charLimit to 0 to disable
func GlobbyHash(charLimit int, dataToHash ...any) string {
	h := NewWeblensHash()

	s := fmt.Sprint(dataToHash...)
	h.Add([]byte(s))

	if charLimit != 0 && charLimit < 16 {
		return h.Done(charLimit)
	} else {
		return h.Done(16)
	}
}

func identifyPanic() string {
	var name, file string
	var line int
	var pc [16]uintptr

	n := runtime.Callers(3, pc[:])
	for _, pc := range pc[:n] {
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		file, line = fn.FileLine(pc)
		name = fn.Name()
		if !strings.HasPrefix(name, "runtime.") || strings.HasPrefix(name, "/opt/homebrew") {
			break
		}
	}

	if file != "" {
		return fmt.Sprintf("%v:%v", filepath.Base(file), line)
	}

	return fmt.Sprintf("pc:%x", pc)
}

func RecoverPanic(preText string) {
	r := recover()
	if r == nil {
		return
	}

	err, ok := r.(error)
	if !ok {
		log.ErrorCatcher.Println(preText, identifyPanic(), r)
	} else {
		log.ErrTrace(err)
	}

}

// OracleReader is almost exactly like io.ReadAll, but if we know how long the content is,
// we can allocate the whole array up front, saving a bit of time (I hope)
func OracleReader(r io.Reader, readerSize int64) ([]byte, error) {
	b := make([]byte, 0, readerSize)
	for {
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return b, err
		}

		if len(b) == cap(b) {
			// Add more capacity (let append pick how much).
			b = append(b, 0)[:len(b)]
		}
	}
}
