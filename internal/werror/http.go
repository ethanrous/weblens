package werror

var ErrBadAuthScheme = NewWeblensError("invalid authorization scheme")
var ErrBasicAuthFormat = NewWeblensError("did not get expected encoded basic auth format")
var ErrEmptyAuth = NewWeblensError("empty auth header not allowed on endpoint")
var ErrCoreOriginate = NewWeblensError("core server attempted to ping remote server")
var ErrNoAddress = NewWeblensError("trying to make request to core without a core address")
var ErrNoKey = NewWeblensError("trying to make request to core without an api key")
var ErrNoBody = NewWeblensError("trying to read http body with no content")
var ErrBodyNotAllowed = NewWeblensError("trying to read http body of GET request")
var ErrCasterDoubleClose = NewWeblensError("trying to close an already disabled caster")
var ErrUnknownWebsocketAction = NewWeblensError("did not recognize websocket action type")
