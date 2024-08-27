package werror

import "errors"

var ErrNoFileAccess = &clientSafeErr{
	realError:  errors.New("user does not have access to file"),
	safeErr:    errors.New("file not found"),
	statusCode: 404,
}

var ErrNoShare = errors.New("share not found")

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

var ErrKeyInUse = errors.New("api key already in use")