package slices

import (
	"strconv"
	"strings"
)

// NatSortCompare compares two strings using natural sort order.
func NatSortCompare(a, b string) int {
	for {
		if p := commonPrefix(a, b); p != 0 {
			a = a[p:]
			b = b[p:]
		}

		if len(a) == 0 {
			return -len(b)
		}

		if ia := digits(a); ia > 0 {
			if ib := digits(b); ib > 0 {
				// Both sides have digits.
				an, aerr := strconv.ParseUint(a[:ia], 10, 64)
				bn, berr := strconv.ParseUint(b[:ib], 10, 64)

				if aerr == nil && berr == nil {
					// Fast path: both fit in uint64
					if an != bn {
						// #nosec G40
						return int(an - bn)
					}
					// Semantically the same digits, e.g. "00" == "0", "01" == "1". In
					// this case, only continue processing if there's trailing data on
					// both sides, otherwise do lexical comparison.
					if ia != len(a) && ib != len(b) {
						a = a[ia:]
						b = b[ib:]

						continue
					}
				} else {
					// Slow path: at least one number exceeds uint64
					// Both are still pure digits (verified by ia > 0 and ib > 0)
					result := compareNumericStrings(a[:ia], b[:ib])
					if result != 0 {
						return result
					}
					// Numbers are semantically equal, continue if both have trailing data
					if ia != len(a) && ib != len(b) {
						a = a[ia:]
						b = b[ib:]

						continue
					}
				}
			}
		}

		return strings.Compare(a, b)
	}
}

// commonPrefix returns the common prefix except for digits.
func commonPrefix(str1, str2 string) int {
	lenLonger := len(str1)
	if lenStr2 := len(str2); lenStr2 < lenLonger {
		lenLonger = lenStr2
	}

	if lenLonger == 0 {
		return 0
	}

	_ = str1[lenLonger-1]
	_ = str2[lenLonger-1]

	for i := 0; i < lenLonger; i++ {
		char1 := str1[i]
		char2 := str2[i]

		if (char1 >= '0' && char1 <= '9') || (char2 >= '0' && char2 <= '9') || char1 != char2 {
			return i
		}
	}

	return lenLonger
}

func digits(s string) int {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return i
		}
	}

	return len(s)
}

// compareNumericStrings compares two numeric strings without parsing them.
// This handles arbitrarily large numbers that don't fit in uint64.
// It does no memory allocation.
func compareNumericStrings(a, b string) int {
	// Strip leading zeros
	a = trimLeadingZeros(a)
	b = trimLeadingZeros(b)

	// Compare by length first (more digits = larger number)
	if len(a) != len(b) {
		return len(a) - len(b)
	}

	// Same length: lexical comparison works correctly for digits
	return strings.Compare(a, b)
}

// trimLeadingZeros removes leading zeros from a numeric string.
// Returns "0" if the string is all zeros.
// This function does no memory allocation (only string slicing).
func trimLeadingZeros(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] != '0' {
			return s[i:]
		}
	}
	// All zeros - return "0"
	if len(s) > 0 {
		return s[len(s)-1:]
	}

	return s
}
