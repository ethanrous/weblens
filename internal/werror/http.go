package werror

import "errors"

var ErrBadAuthScheme = errors.New("invalid authorization scheme")
var ErrBasicAuthFormat = errors.New("did not get expected encoded basic auth format")
var ErrNoBody = errors.New("trying to read http body with no content")
var ErrCasterDoubleClose = errors.New("trying to close an already disabled caster")
