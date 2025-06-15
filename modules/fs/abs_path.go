package fs

import (
	"os"
	"strings"
	"sync"

	"github.com/ethanrous/weblens/modules/errors"
	"github.com/rs/zerolog/log"
)

var absPathMap = make(map[string]string)
var pathMapLock = sync.RWMutex{}

func RegisterAbsolutePrefix(alias, path string) error {
	log.Trace().Msgf("Registering absolute path alias: %s -> %s", alias, path)

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

func getAbsolutePrefix(alias string) (string, error) {
	pathMapLock.RLock()
	defer pathMapLock.RUnlock()

	root, ok := absPathMap[alias]
	if !ok {
		return "", errors.Errorf("no absolute path registered for alias: %s", alias)
	}

	return root, nil
}

func (wf Filepath) ToAbsolute() string {
	if wf.RootAlias == "" {
		return ""
	}

	absPrefix, err := getAbsolutePrefix(wf.RootAlias)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to get absolute prefix for alias: %s", wf.RootAlias)
		return ""
	}

	return absPrefix + wf.RelPath
}
