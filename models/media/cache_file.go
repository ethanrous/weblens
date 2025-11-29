package media

import (
	"fmt"

	"github.com/pkg/errors"
)

// FmtCacheFileName generates the cache file name for the given media, quality, and page number.
func FmtCacheFileName(mId string, quality MediaQuality, pageNum int) (string, error) {
	switch MediaQuality(quality) {
	case LowRes, HighRes:
		break
	default:
		return "", errors.WithStack(ErrInvalidQuality)
	}

	var pageNumStr string
	if pageNum > 1 && quality == HighRes {
		pageNumStr = fmt.Sprintf("_%d", pageNum)
	}

	filename := fmt.Sprintf("%s-%s%s.webp", mId, quality, pageNumStr)

	return filename, nil
}
