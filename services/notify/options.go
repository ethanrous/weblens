package notify

import (
	"github.com/ethanrous/weblens/modules/structs"
)

type FileNotificationOptions struct {
	PreMoveParentId string
	MediaInfo       structs.MediaInfo
}

func consolidateFileOptions(options ...FileNotificationOptions) FileNotificationOptions {
	var consolidated FileNotificationOptions

	for _, opt := range options {
		if opt.PreMoveParentId != "" {
			consolidated.PreMoveParentId = opt.PreMoveParentId
		}
		if opt.MediaInfo.ContentId != "" {
			consolidated.MediaInfo = opt.MediaInfo
		}
	}

	return consolidated

}
