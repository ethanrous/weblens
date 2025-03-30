package fs

import (
	"strings"
	"sync"
)

var absPathMap = make(map[string]string)
var pathMapLock = sync.RWMutex{}

func RegisterAbsolutePrefix(alias, path string) error {
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	pathMapLock.Lock()
	defer pathMapLock.Unlock()
	absPathMap[alias] = path

	return nil
}

func GetAbsolutePrefix(alias string) string {
	pathMapLock.RLock()
	defer pathMapLock.RUnlock()
	return absPathMap[alias]
}

func (wf Filepath) ToAbsolute() string {
	if wf.rootAlias == "" {
		return ""
	}

	absPrefix := GetAbsolutePrefix(wf.rootAlias)
	return absPrefix + wf.relPath
}
