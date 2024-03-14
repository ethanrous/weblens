package dataStore

import (
	"fmt"
	"sync"

	"github.com/ethrousseau/weblens/api/util"
)

var mediaMap map[string]*Media = map[string]*Media{}
var mediaMapLock *sync.Mutex = &sync.Mutex{}

func MediaInit() error {
	ms, err := fddb.getAllMedia()
	if err != nil {
		return err
	}

	for _, m := range ms {
		mediaMapAdd(m)
	}

	return nil
}

func mediaMapAdd(m *Media) {
	if m == nil {
		util.DisplayError(fmt.Errorf("attempt to set nil media in map"))
		return
	}
	if !m.imported {
		util.DisplayError(fmt.Errorf("tried adding non-imported media to map"))
		return
	}

	mediaMapLock.Lock()

	if mediaMap[m.MediaId] != nil {
		mediaMapLock.Unlock()
		util.Error.Println(fmt.Errorf("attempt to re-add media already in map"))
		return
	}

	if m.PageCount == 0 {
		m.PageCount = 1
		err := fddb.UpdateMedia(m)
		if err != nil {
			util.DisplayError(err)
		}
	}

	if m.fullresCacheFiles == nil {
		m.fullresCacheFiles = make([]*WeblensFile, m.PageCount)
	}
	if m.FullresCacheIds == nil {
		m.FullresCacheIds = make([]string, m.PageCount)
	}

	mediaMap[m.MediaId] = m

	mediaMapLock.Unlock()

	orphaned := true
	for _, fId := range m.FileIds {
		f := FsTreeGet(fId)
		if f == nil {
			m.RemoveFile(fId)
			continue
		}
		orphaned = false
		f.SetMedia(m)
	}
	if orphaned {
		removeMedia(m)
	}
}

func MediaMapGet(mId string) (m *Media, err error) {
	mediaMapLock.Lock()
	m = mediaMap[mId]
	mediaMapLock.Unlock()

	if m == nil {
		m = fddb.getMedia(mId)
		if m == nil {
			return m, ErrNoMedia
		}
		m.imported = true
		mediaMapAdd(m)
	}

	return
}

func loadMediaByFile(f *WeblensFile) (err error) {
	var m *Media
	m, err = fddb.getMediaByFile(f)
	if err != nil {
		return
	}
	existingM, err := MediaMapGet(m.Id())
	if existingM != nil {
		m = existingM
	} else if err == ErrNoMedia {
		mediaMapAdd(m)
	} else if err != nil {
		return
	}

	f.SetMedia(m)

	return nil
}

func loadManyMedias(fs []*WeblensFile) (err error) {
	fs = util.Filter(fs, func(f *WeblensFile) bool { return f.media == nil })
	ms, err := fddb.getManyMediasByFiles(fs)
	if err != nil {
		return
	}

	util.Each(ms, func(m *Media) {
		mediaMapAdd(m)
	})

	return
}

func removeMedia(m *Media) {
	f, err := m.getCacheFile(Thumbnail, false, 0)
	if err == nil {
		PermenantlyDeleteFile(f, voidCaster)
	}
	f = nil
	for page := range m.PageCount + 1 {
		f, err = m.getCacheFile(Fullres, false, page)
		if err == nil {
			PermenantlyDeleteFile(f, voidCaster)
		}
	}

	err = fddb.deleteMedia(m.Id())
	if err != nil {
		util.DisplayError(err)
		return
	}

	mediaMapLock.Lock()
	delete(mediaMap, m.Id())
	mediaMapLock.Unlock()
}

func GetRealFile(m *Media) (*WeblensFile, error) {
	for _, fId := range m.FileIds {
		f := FsTreeGet(fId)
		if f != nil {
			return f, nil
		}
	}

	// None of the files that this media uses are present any longer, delete media
	removeMedia(m)
	return nil, ErrNoFile
}

func CleanOrphanedMedias() {
	mediaMapLock.Lock()
	allMedias := util.MapToSlicePure(mediaMap)
	mediaMapLock.Unlock()

	var orphaned []*Media
	for _, m := range allMedias {
		isOrphan := true
		for _, fId := range m.FileIds {
			f := FsTreeGet(fId)
			if f != nil {
				isOrphan = false
				break
			}
		}
		if isOrphan {
			orphaned = append(orphaned, m)
		}
	}

	util.Debug.Println(util.Map(orphaned, func(m *Media) string { return m.Id() }))

	// fddb.deleteManyMedias(util.Map(orphaned, func(m *Media) string {return m.Id()})	)

}
