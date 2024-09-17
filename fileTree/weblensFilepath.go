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
	if path[0] == '/' {
		path = path[1:]
	}

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

func (wf WeblensFilepath) Filename() string {
	filename := wf.relPath
	if len(filename) != 0 && filename[len(filename)-1] == '/' {
		filename = filename[:len(filename)-1]
	}
	return filepath.Base(filename)
}

func (wf WeblensFilepath) IsDir() bool {
	return len(wf.relPath) == 0 || wf.relPath[len(wf.relPath)-1] == '/'
}

func (wf WeblensFilepath) Dir() WeblensFilepath {
	dirPath := wf.relPath
	if len(wf.relPath) != 0 && wf.relPath[len(wf.relPath)-1] == '/' {
		dirPath = dirPath[:len(dirPath)-1]
	}

	return WeblensFilepath{
		rootAlias: wf.rootAlias,
		relPath:   filepath.Dir(dirPath),
	}
}

func (wf WeblensFilepath) Child(childName string, childIsDir bool) WeblensFilepath {
	relPath := filepath.Join(wf.relPath, childName)
	if childIsDir {
		relPath += "/"
	}
	return WeblensFilepath{
		rootAlias: wf.rootAlias,
		relPath:   relPath,
	}
}

func (wf WeblensFilepath) String() string {
	return wf.ToPortable()
}
