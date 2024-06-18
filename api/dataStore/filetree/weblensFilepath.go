package filetree

import (
	"path/filepath"
	"strings"

	"github.com/ethrousseau/weblens/api/types"
)

type WebLensFilepath struct {
	base fpBase
	ext  string
}

type fpBase string

const (
	mediaBase fpBase = "MEDIA"
)

func FilepathFromAbs(absolutePath string, mediaRoot types.WeblensFile) WebLensFilepath {
	ext := strings.TrimPrefix(absolutePath, mediaRoot.GetAbsPath())

	if len(absolutePath) == len(mediaRoot.GetAbsPath()) {
		panic("Abs path is not under mediaService root")
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

func (wf WebLensFilepath) ToAbsPath(mediaRoot types.WeblensFile) string {
	var realBase string
	if wf.base == mediaBase {
		realBase = mediaRoot.GetAbsPath()
	}
	return filepath.Join(realBase, wf.ext)
}

func (wf WebLensFilepath) ToPortable() string {
	return string(wf.base) + ":" + wf.ext
}
