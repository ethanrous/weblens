package fs

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type Filepath struct {
	rootAlias string
	relPath   string
}

func NewFilePath(rootAlias, absolutePath string) (Filepath, error) {
	var root string
	if root = GetAbsolutePrefix(root); root == "" {
		return Filepath{}, errors.Errorf("root alias %s not registered", rootAlias)
	}

	path := strings.TrimPrefix(absolutePath, root)
	if len(path) != 0 && path[0] == '/' {
		path = path[1:]
	}

	return Filepath{
		rootAlias: rootAlias,
		relPath:   path,
	}, nil
}

func ParsePortable(portablePath string) (Filepath, error) {
	colonIndex := strings.Index(portablePath, ":")
	if colonIndex == -1 {
		return Filepath{}, errors.New("invalid portable path format: no colon found")
	}
	prefix := portablePath[:colonIndex]
	postfix := portablePath[colonIndex+1:]
	return Filepath{
		rootAlias: prefix,
		relPath:   postfix,
	}, nil
}

func (wf Filepath) RootName() string {
	return wf.rootAlias
}

func (wf Filepath) OverwriteRoot(newRoot string) Filepath {
	wf.rootAlias = newRoot
	return wf
}

func (wf Filepath) RelativePath() string {
	return wf.relPath
}

func (wf Filepath) ToPortable() string {
	if wf.rootAlias == "" {
		return ""
	}
	return wf.rootAlias + ":" + wf.relPath
}

func (wf Filepath) Filename() string {
	filename := wf.relPath
	if len(filename) != 0 && filename[len(filename)-1] == '/' {
		filename = filename[:len(filename)-1]
	}
	return filepath.Base(filename)
}

func (wf Filepath) IsDir() bool {
	return len(wf.relPath) == 0 || wf.relPath[len(wf.relPath)-1] == '/'
}

// Dir returns the filepath of the directory that the file is in.
func (wf Filepath) Dir() Filepath {
	dirPath := wf.relPath
	if len(wf.relPath) != 0 && wf.relPath[len(wf.relPath)-1] == '/' {
		dirPath = dirPath[:len(dirPath)-1]
	}

	return Filepath{
		rootAlias: wf.rootAlias,
		relPath:   filepath.Dir(dirPath),
	}
}

func (wf Filepath) Child(childName string, childIsDir bool) Filepath {
	relPath := filepath.Join(wf.relPath, childName)
	if childIsDir {
		relPath += "/"
	}
	return Filepath{
		rootAlias: wf.rootAlias,
		relPath:   relPath,
	}
}

func (wf Filepath) String() string {
	return wf.ToPortable()
}
