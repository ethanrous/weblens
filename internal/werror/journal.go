package werror

import "errors"

var ErrNoFileAction = ClientSafeErr{
	realError:  errors.New("could not find file action"),
	safeErr:    nil,
	statusCode: 404,
}

var ErrNoLifetime = ClientSafeErr{
	realError:  errors.New("could not find lifetime with id [%s]"),
	safeErr:    errors.New("could not find lifetime"),
	statusCode: 404,
}

var ErrBadTimestamp = ClientSafeErr{
	realError:  errors.New("a positive timestamp query param is required"),
	safeErr:    nil,
	statusCode: 400,
}

var ErrNoJournal = ClientSafeErr{
	realError:  errors.New("could not load journal"),
	safeErr:    nil,
	statusCode: 500,
}
