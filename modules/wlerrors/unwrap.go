package wlerrors

import "errors"

// Is reports whether any error in err's chain matches target.
func Is(err error, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target, and if so, sets target to that error value and returns true.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Ignore returns nil if err matches any of the errors in the ignore list, otherwise returns err.
func Ignore(err error, ignore ...error) error {
	if err == nil {
		return nil
	}

	for _, i := range ignore {
		if errors.Is(err, i) {
			return nil
		}
	}

	return err
}
