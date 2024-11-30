package werror

import "errors"

var ErrUserNotActive = errors.New("user is not active")
var ErrUserAlreadyExists = clientSafeErr{realError: errors.New("user already exists"), statusCode: 409}

var ErrBadUserOrPass = errors.New("username or password is incorrect")

var ErrBadPassword = clientSafeErr{
	realError:  errors.New("password provided does not authenticate user"),
	safeErr:    ErrBadUserOrPass,
	statusCode: 404,
}

var ErrNoUserLogin = clientSafeErr{
	realError:  errors.New("could not find user to login"),
	safeErr:    ErrBadUserOrPass,
	statusCode: 404,
}

var ErrNoUser = clientSafeErr{
	safeErr:    errors.New("could not find user"),
	statusCode: 404,
}

var ErrCtxMissingUser = clientSafeErr{realError: errors.New("user not found in context"), statusCode: 500}
