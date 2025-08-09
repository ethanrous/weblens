package media

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Regex to capture degrees, minutes, seconds, and direction
var coordRe = regexp.MustCompile(`(\d+)\s*deg\s*(\d+)'\s*([\d\.]+)"\s*([NSEW])`)

// Converts a coordinate string to decimal degrees
func parseCoordinate(coord string) (float64, error) {
	matches := coordRe.FindStringSubmatch(coord)
	if len(matches) != 5 {
		return 0, fmt.Errorf("invalid coordinate format")
	}

	degrees, _ := strconv.ParseFloat(matches[1], 64)
	minutes, _ := strconv.ParseFloat(matches[2], 64)
	seconds, _ := strconv.ParseFloat(matches[3], 64)
	dir := matches[4]

	decimal := degrees + minutes/60 + seconds/3600
	if dir == "S" || dir == "W" {
		decimal = -decimal
	}

	return decimal, nil
}

// Parses a full coordinate string
func getDecimalCoords(coordStr string) (lat, lon float64, err error) {
	parts := strings.Split(coordStr, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("coordinate string must have two parts separated by a comma")
	}

	lat, err = parseCoordinate(strings.TrimSpace(parts[0]))
	if err != nil {
		return
	}

	lon, err = parseCoordinate(strings.TrimSpace(parts[1]))

	return
}
