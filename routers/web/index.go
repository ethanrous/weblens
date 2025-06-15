package web

import (
	"fmt"
	"strings"

	cover_model "github.com/ethanrous/weblens/models/cover"
	media_model "github.com/ethanrous/weblens/models/media"
	share_model "github.com/ethanrous/weblens/models/share"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/services/context"
	"github.com/rs/zerolog/log"
)

func getIndexFields(ctx context.RequestContext, proxyAddress string) (fields indexFields) {
	hasImage := false
	path := ctx.Req.URL.Path

	if path[0] == '/' {
		path = path[1:]
	}

	fields.Url = fmt.Sprintf("%s/%s", proxyAddress, path)

	if strings.HasPrefix(path, "files/share/") {
		path = path[len("files/share/"):]
		slashIndex := strings.Index(path, "/")

		if slashIndex != -1 {
			path = path[:slashIndex]
		}

		shareId := share_model.ShareIdFromString(path)
		ctx.Log().Debug().Msgf("Share ID: %s", shareId)

		share, err := share_model.GetShareById(ctx, shareId)
		if err != nil && errors.Is(err, share_model.ErrShareNotFound) {
			log.Error().Stack().Err(err).Msg("")

			return fields
		}

		if share != nil {
			f, err := ctx.FileService.GetFileById(ctx, share.FileId)
			if err != nil {
				log.Error().Stack().Err(err).Msg("")

				return fields
			}

			// TODO: consider sending file name in private shares. An option, perhaps?
			fields.Title = f.GetPortablePath().Filename() + " - Weblens"

			if !share.IsPublic() {
				fields.Description = "Weblens private file share. Sign in with your weblens account to view"
				fields.Image = "/static/favicon_48x48.png"

				return fields
			}

			fields.Description = "Weblens file share"

			var m *media_model.Media

			if f.IsDir() {
				cover, err := cover_model.GetCoverByFolderId(ctx, f.ID())
				if err == nil {
					m, _ = media_model.GetMediaByContentId(ctx, cover.CoverPhotoId)
				} else {
					hasImage = true
					fields.Image = "/static/folder.png"
				}
			} else {
				m, _ = media_model.GetMediaByContentId(ctx, f.GetContentId())
			}

			if m != nil && !hasImage {
				hasImage = true
				fields.Image = fmt.Sprintf("/api/v1/media/%s.webp?quality=thumbnail&shareId=%s", m.ContentID, share.ID().Hex())
			}
		}
	}

	if !hasImage {
		fields.Image = "/static/logo_1200.png"
	}

	return fields
}

// func getImageUrl() string {
// 	return "/logo_1200.png"
// }
