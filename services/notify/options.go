package notify

import (
	"github.com/ethanrous/weblens/modules/structs"
)

// FileNotificationOptions contains optional parameters for file notifications.
type FileNotificationOptions struct {
	PreMoveParentID string
	MediaInfo       structs.MediaInfo
}

func consolidateFileOptions(options ...FileNotificationOptions) FileNotificationOptions {
	var consolidated FileNotificationOptions

	for _, opt := range options {
		if opt.PreMoveParentID != "" {
			consolidated.PreMoveParentID = opt.PreMoveParentID
		}

		if opt.MediaInfo.ContentID != "" {
			consolidated.MediaInfo = opt.MediaInfo
		}
	}

	return consolidated
}
