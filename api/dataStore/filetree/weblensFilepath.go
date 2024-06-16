package filetree

import (
	"path/filepath"
	"strings"

	"github.com/ethrousseau/weblens/api/dataStore"
)

type WebLensFilepath struct {
	base fpBase
	ext  string
}

type fpBase string

const (
	mediaBase fpBase = "MEDIA"
)

func FilepathFromAbs(absolutePath string) WebLensFilepath {
	ext := strings.TrimPrefix(absolutePath, dataStore.mediaRoot.absolutePath)

	if len(absolutePath) == len(dataStore.mediaRoot.absolutePath) {
		panic("Abs path is not under media root")
	}

	return WebLensFilepath{
		base: mediaBase,
		ext:  ext,
	}
}

func FilepathFromPortable(portablePath string) WebLensFilepath {
	colonIndex := strings.Index(portablePath, "/")
	prefix := portablePath[:colonIndex]
	postfix := portablePath[colonIndex+1:]
	return WebLensFilepath{
		base: fpBase(prefix),
		ext:  postfix,
	}
}

func (wf WebLensFilepath) ToAbsPath() string {
	var realBase string
	if wf.base == mediaBase {
		realBase = dataStore.mediaRoot.absolutePath
	}
	return filepath.Join(realBase, wf.ext)
}

func (wf WebLensFilepath) ToPortable() string {
	return string(wf.base) + ":" + wf.ext
}
