package media

import (
	"github.com/barasher/go-exiftool"
	"github.com/ethrousseau/weblens/api/util"
)

func NewExif(targetSize, currentSize int64, gexift *exiftool.Exiftool) *exiftool.Exiftool {
	if targetSize <= currentSize {
		return gexift
	}
	if gexift != nil {
		err := gexift.Close()
		util.ErrTrace(err)
		gexift = nil
	}
	buf := make([]byte, int(targetSize))
	et, err := exiftool.NewExiftool(exiftool.Api("largefilesupport"),
		exiftool.ExtractAllBinaryMetadata(), exiftool.Buffer(buf, int(targetSize)))
	if err != nil {
		util.ErrTrace(err)
		return nil
	}
	gexift = et

	return gexift
}
