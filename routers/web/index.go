package web

import (
	"fmt"
	"strings"

	cover_model "github.com/ethanrous/weblens/models/cover"
	media_model "github.com/ethanrous/weblens/models/media"
	share_model "github.com/ethanrous/weblens/models/share"
	"github.com/ethanrous/weblens/services/context"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

func getIndexFields(ctx context.RequestContext, proxyAddress string) (fields indexFields) {
	var hasImage bool
	path := ctx.Req.URL.Path

	if path[0] == '/' {
		path = path[1:]
	}
	fields.Url = fmt.Sprintf("%s/%s", proxyAddress, path)

	if strings.HasPrefix(path, "files/share/") {
		path = path[len("files/share/"):]
		slashIndex := strings.Index(path, "/")
		if slashIndex == -1 {
			return fields
		}

		shareId := path[:slashIndex]
		share, err := share_model.GetShareById(ctx, shareId)
		if err != nil && errors.Is(err, share_model.ErrShareNotFound) {
			log.Error().Stack().Err(err).Msg("")
			return fields
		}
		if share != nil {
			if !share.IsPublic() {
				fields.Title = "Sign in to view"
				fields.Description = "Private file share"
				fields.Image = "/logo_1200.png"
				return fields
			}

			f, err := ctx.FileService.GetFileById(share.FileId)
			if err != nil {
				log.Error().Stack().Err(err).Msg("")
				return fields
			}
			if f != nil {
				fields.Title = f.GetPortablePath().Filename()
				fields.Description = "Weblens file share"

				var m *media_model.Media

				if f.IsDir() {
					cover, err := cover_model.GetCoverByFolderId(ctx, f.ID())
					if err == nil {
						m, _ = media_model.GetMediaByContentId(ctx, cover.CoverPhotoId)
					} else {
						imgUrl := fmt.Sprintf("%s/api/static/folder.png", proxyAddress)
						hasImage = true
						fields.Image = imgUrl
					}

				} else {
					m, _ = media_model.GetMediaByContentId(ctx, f.GetContentId())
				}

				if m != nil && !hasImage {
					imgUrl := fmt.Sprintf(
						"%s/api/media/%s.webp?quality=thumbnail&shareId=%s", proxyAddress,
						f.GetContentId(), share.ID(),
					)
					hasImage = true
					fields.Image = imgUrl
				}
			}
		}
	}

	if !hasImage {
		fields.Image = "/logo_1200.png"
	}

	return fields
}

// func getImageUrl() string {
// 	return "/logo_1200.png"
// }
