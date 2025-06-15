package user

import (
	"context"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/modules/errors"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	// ErrUsernameTooShort is returned when the username is too short.
	ErrUsernameTooShort     = errors.Statusf(http.StatusBadRequest, "username is too short")
	ErrUsernameTooLong      = errors.Statusf(http.StatusBadRequest, "username is too long")
	ErrUsernameInvalidChars = errors.Statusf(http.StatusBadRequest, "username contains invalid characters, only alphanumeric, _ and - are allowed")
	ErrUsernameNotAllowed   = errors.Statusf(http.StatusBadRequest, "username is not allowed")

	ErrPasswordTooShort = errors.Statusf(http.StatusBadRequest, "password is too short")
	ErrPasswordNoDigits = errors.Statusf(http.StatusBadRequest, "password must contain at least one digit")
)

var disallowedUsernames = []string{"PUBLIC", "WEBLENS"}

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)

const minUsernameLength = 3
const maxUsernameLength = 25

const minPasswordLength = 6

func validateUsername(ctx context.Context, username string) error {
	if len(username) < minUsernameLength {
		return ErrUsernameTooShort
	}

	if len(username) > maxUsernameLength {
		return ErrUsernameTooShort
	}

	if slices.Contains(disallowedUsernames, username) {
		return ErrUsernameNotAllowed
	}

	if !usernameRe.MatchString(username) {
		return ErrUsernameInvalidChars
	}

	col, err := db.GetCollection(ctx, UserCollectionKey)
	if err != nil {
		return err
	}

	err = col.FindOne(ctx, bson.M{"username": username}).Decode(&User{})
	err = db.WrapError(err, "failed to check if username exists [%s]", username)
	if err == nil {
		return errors.Statusf(http.StatusConflict, "username already exists")
	} else if !db.IsNotFound(err) {
		return err
	}

	return nil
}

func validatePassword(password string) error {
	if len(password) < minPasswordLength {
		return ErrPasswordTooShort
	}

	if !strings.ContainsAny(password, "0123456789") {
		return ErrPasswordNoDigits
	}

	return nil
}
