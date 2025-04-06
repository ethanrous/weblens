package fs

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type Filepath struct {
	RootAlias string
	RelPath   string
}

func BuildFilePath(rootAlias string, relPath ...string) Filepath {
	return Filepath{
		RootAlias: rootAlias,
		RelPath:   filepath.Join(relPath...),
	}
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
		RootAlias: rootAlias,
		RelPath:   path,
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
		RootAlias: prefix,
		RelPath:   postfix,
	}, nil
}

func IsZeroFilepath(wf Filepath) bool {
	return wf.RootAlias == "" && wf.RelPath == ""
}

func (wf Filepath) RootName() string {
	return wf.RootAlias
}

func (wf Filepath) OverwriteRoot(newRoot string) Filepath {
	wf.RootAlias = newRoot
	return wf
}

func (wf Filepath) RelativePath() string {
	return wf.RelPath
}

func (wf Filepath) ToPortable() string {
	if wf.RootAlias == "" {
		return ""
	}
	return wf.RootAlias + ":" + wf.RelPath
}

func (wf Filepath) Filename() string {
	filename := wf.RelPath
	if len(filename) != 0 && filename[len(filename)-1] == '/' {
		filename = filename[:len(filename)-1]
	}
	return filepath.Base(filename)
}

func (wf Filepath) IsDir() bool {
	return len(wf.RelPath) == 0 || wf.RelPath[len(wf.RelPath)-1] == '/'
}

// Dir returns the filepath of the directory that the file is in.
func (wf Filepath) Dir() Filepath {
	dirPath := wf.RelPath
	if len(wf.RelPath) != 0 && wf.RelPath[len(wf.RelPath)-1] == '/' {
		dirPath = dirPath[:len(dirPath)-1]
	}

	return Filepath{
		RootAlias: wf.RootAlias,
		RelPath:   filepath.Dir(dirPath),
	}
}

func (wf Filepath) Child(childName string, childIsDir bool) Filepath {
	relPath := filepath.Join(wf.RelPath, childName)
	if childIsDir {
		relPath += "/"
	}
	return Filepath{
		RootAlias: wf.RootAlias,
		RelPath:   relPath,
	}
}

func (wf Filepath) Ext() string {
	if wf.IsDir() {
		return ""
	}
	return filepath.Ext(wf.RelPath)
}

func (wf Filepath) String() string {
	return wf.ToPortable()
}
func (wf Filepath) MarshalJSON() ([]byte, error) {
	return []byte(wf.ToPortable()), nil
}

func (wf *Filepath) UnmarshalJSON(b []byte) error {
	portable, err := ParsePortable(string(b))
	if err != nil {
		return err
	}

	*wf = portable

	return nil
}
