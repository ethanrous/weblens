// Package media provides media processing and caching functionality.
package media

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/ethanrous/weblens/modules/wlerrors"
)

const cacheFileFormat = "%s-%s%s.webp"
const pageNumExtensionFormat = "_%d"

var cacheFileFormatRegex = regexp.MustCompile(`^([a-zA-Z0-9_-]+)-(thumbnail|fullres)(?:_(\d+))?\.webp$`)

var errInvalidCacheFilename = wlerrors.Errorf("invalid cache file name")

// FmtCacheFileName generates the cache file name for the given media, quality, and page number.
func FmtCacheFileName(mID string, quality Quality, pageNum int) (string, error) {
	switch Quality(quality) {
	case LowRes, HighRes:
		break
	default:
		return "", wlerrors.WithStack(ErrInvalidQuality)
	}

	var pageNumStr string
	if pageNum > 1 && quality == HighRes {
		pageNumStr = fmt.Sprintf(pageNumExtensionFormat, pageNum)
	}

	filename := fmt.Sprintf(cacheFileFormat, mID, quality, pageNumStr)

	return filename, nil
}

// ParseCacheFileName parses the cache file name and returns the media ID, quality, and page number.
func ParseCacheFileName(filename string) (mID string, quality Quality, pageNum int, err error) {
	var qualityStr string

	var pageNumStr string

	matches := cacheFileFormatRegex.FindStringSubmatch(filename)
	if len(matches) < 3 {
		return "", "", 0, wlerrors.WithStack(errInvalidCacheFilename)
	}

	mID = matches[1]
	qualityStr = matches[2]
	pageNumStr = matches[3]

	switch qualityStr {
	case string(LowRes):
		quality = LowRes
	case string(HighRes):
		quality = HighRes

		if pageNumStr != "" {
			pageNum, err = strconv.Atoi(pageNumStr)
			if err != nil {
				return "", "", 0, wlerrors.Errorf("invalid page number in cache file name: %w", err)
			}
		}
	default:
		return "", "", 0, wlerrors.ReplaceStack(ErrInvalidQuality)
	}

	return mID, quality, pageNum, nil
}
