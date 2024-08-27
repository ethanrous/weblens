package werror

import "errors"

var ErrBadPassword = errors.New("password provided does not authenticate user")
var ErrUserAlreadyExists = errors.New("cannot create two users with the same username")
var ErrUserNotFound = errors.New("could not find user")