package tests

import (
	"testing"

	"github.com/ethrousseau/weblens/api/fileTree"
)

func TestFileTree(t *testing.T) {
	tree := fileTree.NewFileTree("~/weblens-test", "MEDIA")

	tree.NewFile(tree.GetRoot(), "my-things", true)
	tree.Add()
}