package werror

import "github.com/pkg/errors"

var ErrNoFileAccess = errors.New("user does not have access to file")
var ErrNoShare = errors.New("could not find share")

var ErrUserNotAuthorized = New("user does not have access the requested resource")
