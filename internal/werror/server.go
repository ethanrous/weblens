package werror

import "errors"

var ErrServerNotInitialized = errors.New("server not initialized")
var ErrServerIsBackup = errors.New("this operation cannot be performed on a backup server")
