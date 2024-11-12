package werror

import "errors"

var ErrUserNotActive = errors.New("user is not active")
var ErrBadPassword = errors.New("password provided does not authenticate user")
var ErrUserAlreadyExists = &clientSafeErr{realError: errors.New("user already exists"), statusCode: 409}
var ErrUserNotFound = errors.New("could not find user")

var ErrCtxMissingUser = &clientSafeErr{realError: errors.New("user not found in context"), statusCode: 500}
