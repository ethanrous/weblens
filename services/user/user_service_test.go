
package user_test

import (
	"testing"

	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/stretchr/testify/assert"
)

func TestUserServiceConstants(t *testing.T) {
	t.Run("UsersTreeKey is defined", func(t *testing.T) {
		assert.NotEmpty(t, string(file_model.UsersTreeKey))
	})

	t.Run("UserTrashDirName is defined", func(t *testing.T) {
		assert.NotEmpty(t, file_model.UserTrashDirName)
	})

	t.Run("UsersRootPath is defined", func(t *testing.T) {
		assert.NotNil(t, file_model.UsersRootPath)
	})
}

func TestUserHomePathConstruction(t *testing.T) {
	t.Run("constructs user home path", func(t *testing.T) {
		username := "testuser"
		homePath := file_model.UsersRootPath.Child(username, true)

		assert.NotNil(t, homePath)
	})

	t.Run("constructs user trash path", func(t *testing.T) {
		username := "testuser"
		homePath := file_model.UsersRootPath.Child(username, true)
		trashPath := homePath.Child(file_model.UserTrashDirName, true)

		assert.NotNil(t, trashPath)
	})
}
