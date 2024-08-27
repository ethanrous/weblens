package werror

import "errors"

var ErrBadAuthScheme = errors.New("invalid authorization scheme")
var ErrBasicAuthFormat = errors.New("did not get expected encoded basic auth format")
var ErrEmptyAuth = errors.New("empty auth header not allowed on endpoint")
var ErrCoreOriginate = errors.New("core server attempted to ping remote server")
var ErrNoAddress = errors.New("trying to make request to core without a core address")
var ErrNoKey = errors.New("trying to make request to core without an api key")
var ErrNoBody = errors.New("trying to read http body with no content")
var ErrBodyNotAllowed = errors.New("trying to read http body of GET request")
var ErrCasterDoubleClose = errors.New("trying to close an already disabled caster")
var ErrUnknownWebsocketAction = errors.New("did not recognize websocket action type")
