package werror

import (
	"errors"
	"net/http"
)

var ErrUserNotActive = errors.New("user is not active")
var ErrUserAlreadyExists = ClientSafeErr{realError: errors.New("user already exists"), statusCode: 409}

var ErrBadUserOrPass = errors.New("username or password is incorrect")

var ErrBadPassword = ClientSafeErr{
	realError:  errors.New("password provided does not authenticate user"),
	safeErr:    ErrBadUserOrPass,
	statusCode: 401,
}

var ErrNoUserLogin = ClientSafeErr{
	realError:  errors.New("could not find user to login"),
	safeErr:    ErrBadUserOrPass,
	statusCode: http.StatusUnauthorized,
}

var ErrNoUser = ClientSafeErr{
	safeErr:    errors.New("could not find user"),
	statusCode: http.StatusUnauthorized,
}

var ErrCtxMissingUser = ClientSafeErr{realError: errors.New("user not found in context"), statusCode: 500}
