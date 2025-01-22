package werror

import "errors"

var ErrUserNotActive = errors.New("user is not active")
var ErrUserAlreadyExists = ClientSafeErr{realError: errors.New("user already exists"), statusCode: 409}

var ErrBadUserOrPass = errors.New("username or password is incorrect")

var ErrBadPassword = ClientSafeErr{
	realError:  errors.New("password provided does not authenticate user"),
	safeErr:    ErrBadUserOrPass,
	statusCode: 404,
}

var ErrNoUserLogin = ClientSafeErr{
	realError:  errors.New("could not find user to login"),
	safeErr:    ErrBadUserOrPass,
	statusCode: 404,
}

var ErrNoUser = ClientSafeErr{
	safeErr:    errors.New("could not find user"),
	statusCode: 404,
}

var ErrCtxMissingUser = ClientSafeErr{realError: errors.New("user not found in context"), statusCode: 500}
