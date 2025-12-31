// Package log provides colorized logging utilities for HTTP requests and responses.
package log

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

var (
	// RED is the ANSI escape code for red text.
	RED = "\u001b[31m"
	// BLUE is the ANSI escape code for green text.
	BLUE = "\u001b[32m"
	// YELLOW is the ANSI escape code for yellow text.
	YELLOW = "\u001b[33m"
	// ORANGE is the ANSI escape code for cyan text.
	ORANGE = "\u001b[36m"
	// RESET is the ANSI escape code to reset text formatting.
	RESET = "\u001B[0m"
)

// ColorStatus returns an HTTP status code string with color formatting based on the status value.
func ColorStatus(status int) string {
	if status == 0 {
		return RED + strconv.Itoa(status) + RESET
	} else if status < 400 {
		return BLUE + strconv.Itoa(status) + RESET
	} else if status >= 400 && status < 500 {
		return YELLOW + strconv.Itoa(status) + RESET
	} else if status >= 500 {
		return RED + strconv.Itoa(status) + RESET
	}

	return RED + strconv.Itoa(status) + RESET + " UNKNOWN STATUS CODE"
}

// ColorTime returns a duration string with color formatting based on the duration length.
func ColorTime(dur time.Duration) string {
	durString := dur.String()

	lastDigitIndex := strings.LastIndexFunc(
		durString, func(r rune) bool {
			return r < 58
		},
	)
	if len(durString) > 7 {
		durString = (durString[:lastDigitIndex+1])[:4] + durString[lastDigitIndex:]
	}

	if dur < time.Millisecond*200 {
		return durString
	} else if dur < time.Second {
		return YELLOW + durString + RESET
	}

	return RED + durString + RESET
}

// RouteColor returns the route pattern for the request, highlighting unknown routes in red.
func RouteColor(r *http.Request) string {
	route := chi.RouteContext(r.Context()).RoutePattern()

	if route == "/api/v1/*" {
		route = RED + "?" + r.URL.Path + RESET
	}

	return route
}
