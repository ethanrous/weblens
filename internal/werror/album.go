package werror

import "errors"

var albumNotFound = errors.New("album not found")

var ErrNoAlbum = &clientSafeErr{
	safeErr:    albumNotFound,
	statusCode: 404,
}

var ErrNoAlbumAccess = &clientSafeErr{
	realError:  errors.New("user does not have access to album"),
	safeErr:    albumNotFound,
	statusCode: 404,
}
