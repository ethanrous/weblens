package fileTree

import (
	"path/filepath"
	"strings"
)

type WeblensFilepath struct {
	rootAlias string
	relPath   string
}

func NewFilePath(root, rootAlias, absolutePath string) WeblensFilepath {
	path := strings.TrimPrefix(absolutePath, root)

	return WeblensFilepath{
		rootAlias: rootAlias,
		relPath:   path,
	}
}

func ParsePortable(portablePath string) WeblensFilepath {
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
		rootAlias: prefix,
		relPath:   postfix,
	}
}

func (wf WeblensFilepath) RootName() string {
	return wf.rootAlias
}

func (wf WeblensFilepath) RelativePath() string {
	return wf.relPath
}

func (wf WeblensFilepath) ToPortable() string {
	return string(wf.rootAlias) + ":" + wf.relPath
}

func (wf WeblensFilepath) Child(childName string) WeblensFilepath {
	return WeblensFilepath{
		rootAlias: wf.rootAlias,
		relPath:   filepath.Join(wf.relPath, childName),
	}
}

func (wf WeblensFilepath) String() string {
	return wf.ToPortable()
}
