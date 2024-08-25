package werror

import "github.com/pkg/errors"

var ErrBadPassword = errors.New("password provided does not authenticate user")
var ErrUserAlreadyExists = errors.New("cannot create two users with the same username")
