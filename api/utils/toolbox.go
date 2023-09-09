package util

import (
	"fmt"
)

func BoolFromString(input string, emptyIsFalse bool) bool {
	if input == "1" || input == "true" {
		return true
	} else if input == "0" || input == "false" {
		return false
	} else if input == "" {
		return !emptyIsFalse
	} else {
		panic(fmt.Errorf("unexpected boolean string. Unable to determine truthiness of %s", input))
	}
}

func IntFromBool( input bool ) int {
	if input {
		return 1
	} else {
		return 0
	}
}

func SliceRemove(s []any, i int) []any {
    s[i] = s[len(s)-1]
    return s[:len(s)-1]
}
