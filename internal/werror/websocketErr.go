package werror

import "errors"

var ErrSubscriptionNotFound = errors.New("subscription not found")
var ErrNoSubKey = errors.New("subscription key was not provided, but one was expected")
