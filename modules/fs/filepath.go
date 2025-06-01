package fs

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

var ErrInvalidPortablePath = errors.New("invalid portable path format")

type Filepath struct {
	RootAlias string
	RelPath   string
}

func BuildFilePath(rootAlias string, relPath ...string) Filepath {
	path := filepath.Join(relPath...)

	if len(relPath) != 0 {
		last := relPath[len(relPath)-1]
		if len(last) != 0 && last[len(last)-1] == '/' {
			path += "/"
		}
	}

	return Filepath{
		RootAlias: rootAlias,
		RelPath:   path,
	}
}

func NewFilePath(rootAlias, absolutePath string) (Filepath, error) {
	root, err := getAbsolutePrefix(rootAlias)
	if err != nil {
		return Filepath{}, err
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
		return Filepath{}, ErrInvalidPortablePath
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

func (wf Filepath) IsZero() bool {
	return wf.RootAlias == "" && wf.RelPath == ""
}

func (wf Filepath) Depth() int {
	if wf.IsZero() {
		return 0
	}

	if wf.IsRoot() {
		return 1
	}

	return strings.Count(wf.RelPath, "/") + 1
}

func (wf Filepath) IsRoot() bool {
	return wf.RootAlias != "" && wf.RelPath == ""
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

	dirPath = filepath.Dir(dirPath)
	if dirPath == "." {
		dirPath = ""
	} else {
		dirPath += "/"
	}

	return Filepath{
		RootAlias: wf.RootAlias,
		RelPath:   dirPath,
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

func (wf Filepath) ReplacePrefix(prefixPath, newPrefix Filepath) (Filepath, error) {
	if !strings.HasPrefix(wf.RelPath, prefixPath.RelPath) {
		return Filepath{}, errors.Errorf("prefix %s not found in path %s", prefixPath, wf)
	}

	newRelPath := newPrefix.RelPath + strings.TrimPrefix(wf.RelPath, prefixPath.RelPath)

	return Filepath{
		RootAlias: newPrefix.RootAlias,
		RelPath:   newRelPath,
	}, nil
}

func (wf Filepath) String() string {
	return wf.ToPortable()
}

func (wf Filepath) MarshalJSON() ([]byte, error) {
	portable := wf.ToPortable()
	bs, err := json.Marshal(portable)

	return bs, err
}

func (wf *Filepath) UnmarshalJSON(b []byte) error {
	path := ""
	if err := json.Unmarshal(b, &path); err != nil {
		return err
	}
	log.FromContext(context.TODO()).Debug().Msgf("UnmarshalJSON: %s", path)

	portable, err := ParsePortable(path)
	if err != nil {
		return err
	}

	*wf = portable

	return nil
}

func (wf Filepath) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if wf.IsZero() {
		return bson.TypeNull, nil, nil
	}

	portable := wf.ToPortable()

	return bson.MarshalValue(portable)
}

func (wf *Filepath) UnmarshalBSONValue(_ bsontype.Type, data []byte) error {
	if len(data) == 0 {
		return nil
	}

	var target string

	err := bson.UnmarshalValue(bson.TypeString, data, &target)
	if err != nil {
		return err
	}

	portable, err := ParsePortable(target)
	if err != nil {
		return err
	}

	*wf = portable

	return nil
}
