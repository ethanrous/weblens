package file

import (
	"context"
	"fmt"

	file_model "github.com/ethanrous/weblens/models/file"
	user_model "github.com/ethanrous/weblens/models/user"
	file_system "github.com/ethanrous/weblens/modules/fs"
	"github.com/ethanrous/weblens/modules/errors"
)

// GetFileOwner retrieves the user who owns the specified file.
func GetFileOwner(ctx context.Context, file *file_model.WeblensFileImpl) (*user_model.User, error) {
	username, err := file_model.GetFileOwnerName(ctx, file)
	if err != nil {
		return nil, err
	}

	return user_model.GetUserByUsername(ctx, username)
}

const maxDupeCount = 100

// MakeUniqueChildName generates a unique filename within the parent directory by appending a number suffix if necessary.
func MakeUniqueChildName(parent file_system.Filepath, childName string, childIsDir bool) (childPath file_system.Filepath, err error) {
	dupeCount := 0

	if !exists(parent) {
		return childPath, errors.New("parent does not exist")
	}

	// Check if the child already exists
	childPath = parent.Child(childName, childIsDir)
	for exists(childPath) {
		dupeCount++
		tmpName := fmt.Sprintf("%s (%d)", childName, dupeCount)
		childPath = parent.Child(tmpName, childIsDir)

		if dupeCount > maxDupeCount {
			return childPath, errors.New("too many duplicates")
		}
	}

	return childPath, nil
}
