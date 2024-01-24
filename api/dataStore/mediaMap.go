package dataStore

import "sync"

var mediaMap map[string]*Media = map[string]*Media{}
var mediaMapLock *sync.Mutex = &sync.Mutex{}

func mediaMapAdd(m *Media) {
	mediaMapLock.Lock()
	mediaMap[m.MediaId] = m
	mediaMapLock.Unlock()

	FsTreeGet(m.FileId).SetMedia(m)
}

func mediaMapGet(mId string) *Media {
	mediaMapLock.Lock()
	defer mediaMapLock.Unlock()
	return mediaMap[mId]
}
