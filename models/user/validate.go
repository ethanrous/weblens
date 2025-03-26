package user

import (
	"regexp"
	"slices"

	"github.com/pkg/errors"
)

var (
	// ErrUsernameTooShort is returned when the username is too short
	ErrUsernameTooShort     = errors.New("username is too short")
	ErrUsernameInvalidChars = errors.New("username contains invalid characters, only alphanumeric, _ and - are allowed")
	ErrUsernameNotAllowed   = errors.New("username is not allowed")

	ErrPasswordTooShort = errors.New("password is too short")
)

var disallowedUsernames = []string{"PUBLIC", "WEBLENS"}

var thing = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)

func validateUsername(username string) error {
	if len(username) < 3 {
		return ErrUsernameTooShort
	}

	if slices.Contains(disallowedUsernames, username) {
		return ErrUsernameNotAllowed
	}

	if !thing.MatchString(username) {
		return ErrUsernameInvalidChars
	}

	return nil
}

func validatePassword(password string) error {
	if len(password) < 6 {
		return ErrPasswordTooShort
	}

	return nil
}
