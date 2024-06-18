package history

import "github.com/ethrousseau/weblens/api/types"

type lifetime struct {
	fileId    types.FileId
	contentId types.ContentId
}

func (l lifetime) GetFileId() types.FileId {
	return l.fileId
}

func (l lifetime) GetContentId() types.ContentId {
	return l.contentId
}
