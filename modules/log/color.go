package log

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

var (
	RED    = "\u001b[31m"
	BLUE   = "\u001b[32m"
	YELLOW = "\u001b[33m"
	ORANGE = "\u001b[36m"
	RESET  = "\u001B[0m"
)

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
	} else {
		return RED + durString + RESET
	}
}

func RouteColor(r *http.Request) string {
	route := chi.RouteContext(r.Context()).RoutePattern()

	if route == "/api/v1/*" {
		route = RED + "?" + r.URL.Path + RESET
	}

	return route
}
