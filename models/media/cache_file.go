// Package media provides media processing and caching functionality.
package media

import (
	"fmt"

	"github.com/pkg/errors"
)

// FmtCacheFileName generates the cache file name for the given media, quality, and page number.
func FmtCacheFileName(mID string, quality Quality, pageNum int) (string, error) {
	switch Quality(quality) {
	case LowRes, HighRes:
		break
	default:
		return "", errors.WithStack(ErrInvalidQuality)
	}

	var pageNumStr string
	if pageNum > 1 && quality == HighRes {
		pageNumStr = fmt.Sprintf("_%d", pageNum)
	}

	filename := fmt.Sprintf("%s-%s%s.webp", mID, quality, pageNumStr)

	return filename, nil
}
