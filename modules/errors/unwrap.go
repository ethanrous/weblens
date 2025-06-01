package errors

import "errors"

func Is(err error, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target any) bool {
	return errors.As(err, target)
}

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
