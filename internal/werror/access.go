package werror

import (
	"errors"
	"net/http"
)

var genericNotFound = errors.New("404 Not Found")
var ErrNoAuth = errors.New("no auth header provided")

var ErrUserNotAuthorized = ClientSafeErr{
	realError:  errors.New("user does not have access the requested resource"),
	safeErr:    genericNotFound,
	statusCode: http.StatusNotFound,
}

var ErrBadAuth = ClientSafeErr{
	safeErr:    errors.New("recieved non-empty authorization but format was incorrect"),
	statusCode: http.StatusBadRequest,
}

var ErrKeyInUse = ClientSafeErr{
	safeErr:    errors.New("api key already in use"),
	statusCode: http.StatusConflict,
}

var ErrKeyNotFound = ClientSafeErr{
	safeErr:    errors.New("api key was not found"),
	statusCode: http.StatusNotFound,
}

var ErrInvalidToken  = ClientSafeErr{
	safeErr:    errors.New("session token is invalid"),
	statusCode: http.StatusUnauthorized,
}

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

var ErrNoServerInContext = ClientSafeErr{
	safeErr:    errors.New("the requester was expected to identify itself as a server, but no server id was found"),
	statusCode: http.StatusUnauthorized,
}
