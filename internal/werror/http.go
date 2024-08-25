package werror

var ErrBadAuthScheme = New("invalid authorization scheme")
var ErrBasicAuthFormat = New("did not get expected encoded basic auth format")
var ErrEmptyAuth = New("empty auth header not allowed on endpoint")
var ErrCoreOriginate = New("core server attempted to ping remote server")
var ErrNoAddress = New("trying to make request to core without a core address")
var ErrNoKey = New("trying to make request to core without an api key")
var ErrNoBody = New("trying to read http body with no content")
var ErrBodyNotAllowed = New("trying to read http body of GET request")
var ErrCasterDoubleClose = New("trying to close an already disabled caster")
var ErrUnknownWebsocketAction = New("did not recognize websocket action type")
