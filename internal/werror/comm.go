package werror

import "errors"

var ErrNoAuth = errors.New("no auth header provided")
var ErrBadAuth = errors.New("auth header format is unexpected")

var ErrBadAuthScheme = errors.New("invalid authorization scheme")
var ErrBasicAuthFormat = errors.New("did not get expected encoded basic auth format")
var ErrNoBody = errors.New("trying to read http body with no content")
var ErrCasterDoubleClose = errors.New("trying to close an already disabled caster")
