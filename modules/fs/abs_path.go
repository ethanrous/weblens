package fs

import (
	"os"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var absPathMap = make(map[string]string)
var pathMapLock = sync.RWMutex{}

func RegisterAbsolutePrefix(alias, path string) error {
	log.Debug().Msgf("Registering absolute path alias: %s -> %s", alias, path)

	if !strings.HasPrefix(path, "/") {
		return errors.New("absolute path must start with /")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	pathMapLock.Lock()
	defer pathMapLock.Unlock()

	absPathMap[alias] = path

	return nil
}

func getAbsolutePrefix(alias string) string {
	pathMapLock.RLock()
	defer pathMapLock.RUnlock()
	root, ok := absPathMap[alias]
	if !ok {
		panic("No absolute path registered for alias: " + alias)
	}
	return root

}

func (wf Filepath) ToAbsolute() string {
	if wf.RootAlias == "" {
		return ""
	}

	absPrefix := getAbsolutePrefix(wf.RootAlias)
	return absPrefix + wf.RelPath
}
