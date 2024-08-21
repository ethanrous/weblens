package fileTree

import (
	"path/filepath"
	"strings"

	"github.com/ethrousseau/weblens/api/types"
)

type WeblensFilepath struct {
	base string
	ext  string
}

func FilepathFromAbs(absolutePath string) WeblensFilepath {

	rootPath := types.SERV.FileTree.GetRoot().GetAbsPath()
	ext := strings.TrimPrefix(absolutePath, rootPath)

	return WeblensFilepath{
		// TODO
		base: "",
		ext:  ext,
	}
}

func FilepathFromPortable(portablePath string) WeblensFilepath {
	colonIndex := strings.Index(portablePath, ":")
	if colonIndex == -1 {
		// util.ShowErr(
		// 	types.WeblensErrorMsg(
		// 		fmt.Sprintf(
		// 			"could not get colon index parsing portable path [%s]",
		// 			portablePath,
		// 		),
		// 	),
		// )
		return WeblensFilepath{}
	}
	prefix := portablePath[:colonIndex]
	postfix := portablePath[colonIndex+1:]
	return WeblensFilepath{
		base: prefix,
		ext:  postfix,
	}
}

func (wf WeblensFilepath) ToAbsPath() string {
	mediaRoot := types.SERV.FileTree.GetRoot()
	var realBase string
	// TODO
	if wf.base == "" {
		realBase = mediaRoot.GetAbsPath()
	}
	return filepath.Join(realBase, wf.ext)
}

func (wf WeblensFilepath) ToPortable() string {
	return string(wf.base) + ":" + wf.ext
}

func (wf WeblensFilepath) String() string {
	return wf.ToPortable()
}
