package dataStore

import (
	"sync"

	"github.com/ethrousseau/weblens/api/util"
)

var mediaMap map[string]*Media = map[string]*Media{}
var mediaMapLock *sync.Mutex = &sync.Mutex{}

func mediaMapAdd(m *Media) {
	mediaMapLock.Lock()
	if mediaMap[m.MediaId] != nil {
		mediaMapLock.Unlock()
		util.Warning.Println("Attempt to re-add media already in map")
		return
	}
	mediaMap[m.MediaId] = m
	mediaMapLock.Unlock()

	FsTreeGet(m.FileId).SetMedia(m)
}

func MediaMapGet(mId string) *Media {
	mediaMapLock.Lock()
	defer mediaMapLock.Unlock()
	return mediaMap[mId]
}

func loadMediaByFile(f *WeblensFile) (err error) {
	var m *Media
	m, err = fddb.getMediaByFile(f)
	if err != nil {
		return
	}

	mediaMapAdd(m)
	return
}

func removeMediaByFile(f *WeblensFile) (err error) {
	var m *Media
	m, err = f.GetMedia()
	if m == nil || err != nil {
		return
	}
	err = fddb.removeMediaByFile(f)
	if err != nil {
		return
	}

	mediaMapLock.Lock()
	delete(mediaMap, m.Id())
	mediaMapLock.Unlock()
	return
}
