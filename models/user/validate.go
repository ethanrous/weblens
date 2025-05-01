package user

import (
	"net/http"
	"regexp"
	"slices"

	"github.com/ethanrous/weblens/modules/werror"
)

var (
	// ErrUsernameTooShort is returned when the username is too short.
	ErrUsernameTooShort     = werror.Statusf(http.StatusBadRequest, "username is too short")
	ErrUsernameInvalidChars = werror.Statusf(http.StatusBadRequest, "username contains invalid characters, only alphanumeric, _ and - are allowed")
	ErrUsernameNotAllowed   = werror.Statusf(http.StatusBadRequest, "username is not allowed")

	ErrPasswordTooShort = werror.Statusf(http.StatusBadRequest, "password is too short")
)

var disallowedUsernames = []string{"PUBLIC", "WEBLENS"}

var thing = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)

const minUsernameLength = 3
const minPasswordLength = 6

func validateUsername(username string) error {
	if len(username) < minUsernameLength {
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
	if len(password) < minPasswordLength {
		return ErrPasswordTooShort
	}

	return nil
}
