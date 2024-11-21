package werror

import "errors"

var ErrUserNotActive = errors.New("user is not active")
var ErrBadPassword = errors.New("password provided does not authenticate user")
var ErrUserAlreadyExists = &clientSafeErr{realError: errors.New("user already exists"), statusCode: 409}

var ErrNoUser = &clientSafeErr{
	safeErr:    errors.New("could not find user"),
	statusCode: 404,
}

var ErrCtxMissingUser = &clientSafeErr{realError: errors.New("user not found in context"), statusCode: 500}
