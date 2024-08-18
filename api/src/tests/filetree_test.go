package tests

import (
	"testing"

	"github.com/ethrousseau/weblens/api/dataStore/filetree"
)

func TestFileTree(t *testing.T) {
	tree := filetree.NewFileTree("~/weblens-test", "MEDIA")

	tree.NewFile(tree.GetRoot(), "my-things", true)
	tree.Add()
}