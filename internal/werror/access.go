package werror

import (
	"errors"
	"net/http"
)

var genericNotFound = errors.New("404 Not Found")

var ErrUserNotAuthorized = ClientSafeErr{
	realError:  errors.New("user does not have access the requested resource"),
	safeErr:    genericNotFound,
	statusCode: http.StatusNotFound,
}

var ErrKeyInUse = ClientSafeErr{
	safeErr:    errors.New("api key already in use"),
	statusCode: http.StatusConflict,
}

var ErrKeyNotFound = ClientSafeErr{
	realError:  errors.New("api was not found"),
	safeErr:    genericNotFound,
	statusCode: http.StatusNotFound,
}

var ErrInvalidToken = errors.New("session token is invalid")

var ErrTokenExpired = ClientSafeErr{
	safeErr:    errors.New("session token is expired"),
	statusCode: http.StatusUnauthorized,
}

var ErrKeyAlreadyExists = errors.New("api key already exists")
var ErrKeyNoServer = errors.New("api key is not associated with a server")

var ErrNotAdmin = ClientSafeErr{
	safeErr:    errors.New("user must be admin to access this resource"),
	statusCode: http.StatusForbidden,
}

var ErrNotOwner = ClientSafeErr{
	safeErr:    errors.New("user must be server owner to access this resource"),
	statusCode: http.StatusForbidden,
}

var ErrNoPublicUser = ClientSafeErr{
	safeErr:    errors.New("user must be logged in to access this resource"),
	statusCode: http.StatusUnauthorized,
}
