package werror

import "errors"

var ErrNoAlbum = &clientSafeErr{
	safeErr:    errors.New("album not found"),
	statusCode: 404,
}
