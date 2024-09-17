package werror

import "errors"

var ErrNoFileAction = &clientSafeErr{
	realError:  errors.New("could not find file action"),
	safeErr:    nil,
	statusCode: 404,
}

var ErrNoLifetime = &clientSafeErr{
	realError:  errors.New("could not find lifetime"),
	safeErr:    nil,
	statusCode: 404,
}
