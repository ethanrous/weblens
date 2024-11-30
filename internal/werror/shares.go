package werror

import (
	"errors"
	"net/http"
)

var ErrEmptyShare = errors.New("share does not expand any permissions")
var ErrNoShare = errors.New("share not found")
var ErrBadShareType = errors.New("share has a different type than was expected")
var ErrExpectedShareMissing = errors.New("share that was expected to exist was not found")

var ErrNoShareAccess = clientSafeErr{
	realError:  errors.New("user does not have access to share"),
	safeErr:    ErrNoShare,
	statusCode: http.StatusNotFound,
}

var ErrShareAlreadyExists = clientSafeErr{
	safeErr:    errors.New("share already exists"),
	statusCode: 409,
}
