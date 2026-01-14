package user

import (
	"context"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/ethanrous/weblens/models/db"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	// ErrUsernameTooShort is returned when the username is too short.
	ErrUsernameTooShort = wlerrors.Statusf(http.StatusBadRequest, "username is too short")
	// ErrUsernameTooLong is returned when the username exceeds the maximum length.
	ErrUsernameTooLong = wlerrors.Statusf(http.StatusBadRequest, "username is too long")
	// ErrUsernameInvalidChars is returned when the username contains invalid characters.
	ErrUsernameInvalidChars = wlerrors.Statusf(http.StatusBadRequest, "username contains invalid characters, only alphanumeric, _ and - are allowed")
	// ErrUsernameNotAllowed is returned when the username is reserved or disallowed.
	ErrUsernameNotAllowed = wlerrors.Statusf(http.StatusBadRequest, "username is not allowed")

	// ErrPasswordTooShort is returned when the password is too short.
	ErrPasswordTooShort = wlerrors.Statusf(http.StatusBadRequest, "password is too short")
	// ErrPasswordNoDigits is returned when the password contains no digits.
	ErrPasswordNoDigits = wlerrors.Statusf(http.StatusBadRequest, "password must contain at least one digit")
)

var disallowedUsernames = []string{"PUBLIC", "WEBLENS"}

var usernameRe = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)

const minUsernameLength = 2
const maxUsernameLength = 25

const minPasswordLength = 6

// ValidateUser checks if the user's username and password meet the required criteria.
func ValidateUser(ctx context.Context, user *User) error {
	if err := validateUsername(ctx, user.Username); err != nil {
		return err
	}

	if err := validatePassword(user.Password); err != nil {
		return err
	}

	return nil
}

func validateUsername(ctx context.Context, username string) error {
	if len(username) < minUsernameLength {
		return wlerrors.ReplaceStack(ErrUsernameTooShort)
	}

	if len(username) > maxUsernameLength {
		return wlerrors.ReplaceStack(ErrUsernameTooLong)
	}

	if slices.Contains(disallowedUsernames, username) {
		return wlerrors.ReplaceStack(ErrUsernameNotAllowed)
	}

	if !usernameRe.MatchString(username) {
		return wlerrors.Wrap(ErrUsernameInvalidChars, "")
	}

	col, err := db.GetCollection[any](ctx, UserCollectionKey)
	if err != nil {
		return err
	}

	err = col.FindOne(ctx, bson.M{"username": username}).Decode(&User{})

	err = db.WrapError(err, "failed to check if username exists [%s]", username)
	if err == nil {
		return wlerrors.Statusf(http.StatusConflict, "username already exists")
	} else if !db.IsNotFound(err) {
		return err
	}

	return nil
}

func validatePassword(password string) error {
	if len(password) < minPasswordLength {
		return wlerrors.ReplaceStack(ErrPasswordTooShort)
	}

	if !strings.ContainsAny(password, "0123456789") {
		return wlerrors.ReplaceStack(ErrPasswordNoDigits)
	}

	return nil
}
