package werror

import "errors"

var ErrNoFileAccess = &clientSafeErr{
	realError:  errors.New("user does not have access to file"),
	safeErr:    errors.New("file does not exist or user does not have access to it"),
	statusCode: 404,
}

var ErrNoShare = errors.New("share not found")
var ErrExpectedShareMissing = errors.New("Share that was expected to exist was not found")
var ErrBadShareType = errors.New("Share has a different type than was expected")

var ErrNoShareAccess = &clientSafeErr{
	realError:  errors.New("user does not have access to share"),
	safeErr:    ErrNoShare,
	statusCode: 404,
}

var ErrUserNotAuthorized = &clientSafeErr{
	realError:  errors.New("user does not have access the requested resource"),
	safeErr:    errors.New("resource not found"),
	statusCode: 404,
}

var ErrKeyInUse = &clientSafeErr{
	realError:  errors.New("api key already in use"),
	statusCode: 400,
}

var ErrKeyNotFound = errors.New("api was not found")
var ErrInvalidToken = errors.New("session token is expired or invalid")
var ErrKeyAlreadyExists = errors.New("api key already exists")
var ErrKeyNoServer = errors.New("api key is not associated with a server")
