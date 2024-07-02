package filetree

import (
	"path/filepath"
	"strings"

	"github.com/ethrousseau/weblens/api/types"
)

type weblensFilepath struct {
	base fpBase
	ext  string
}

type fpBase string

const (
	mediaBase fpBase = "MEDIA"
)

func FilepathFromAbs(absolutePath string) types.WeblensFilepath {

	rootPath := types.SERV.FileTree.GetRoot().GetAbsPath()
	ext := strings.TrimPrefix(absolutePath, rootPath)

	if len(absolutePath) == len(rootPath) {
		panic("Abs path is not under mediaService root")
	}

	return weblensFilepath{
		base: mediaBase,
		ext:  ext,
	}
}

func FilepathFromPortable(portablePath string) types.WeblensFilepath {
	colonIndex := strings.Index(portablePath, "/")
	prefix := portablePath[:colonIndex]
	postfix := portablePath[colonIndex+1:]
	return weblensFilepath{
		base: fpBase(prefix),
		ext:  postfix,
	}
}

func (wf weblensFilepath) ToAbsPath() string {
	mediaRoot := types.SERV.FileTree.GetRoot()
	var realBase string
	if wf.base == mediaBase {
		realBase = mediaRoot.GetAbsPath()
	}
	return filepath.Join(realBase, wf.ext)
}

func (wf weblensFilepath) ToPortable() string {
	return string(wf.base) + ":" + wf.ext
}
