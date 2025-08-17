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

const apiBasePath = "/api/v1"

func getIndexFields(ctx context.RequestContext, proxyAddress string) (fields indexFields) {
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
					fields.Image = "/static/folder.png"
				}
			} else {
				m, _ = media_model.GetMediaByContentId(ctx, f.GetContentId())
			}

			if m != nil && fields.Image == "" {
				fields.Image = fmt.Sprintf("%s/media/%s.webp?quality=thumbnail&shareId=%s", apiBasePath, m.ContentID, share.ID().Hex())
			}
		}
	} else if strings.HasPrefix(path, "media/") {
		handleMediaPage(ctx, path[len("media/"):], &fields)
	}
	ctx.Log().Debug().Msgf("Media path?: %s", path)

	if fields.Image == "" {
		fields.Image = "/static/favicon_48x48.png"
	}

	return fields
}

func handleMediaPage(ctx context.RequestContext, mediaId string, fields *indexFields) {
	m, err := media_model.GetMediaByContentId(ctx, mediaId)
	if err != nil {
		log.Error().Stack().Err(err).Msg("")

		return
	}

	if m == nil {
		return
	}

	fields.Description = "Weblens Media"
	fields.Image = fmt.Sprintf("%s/media/%s.webp?quality=thumbnail", apiBasePath, m.ContentID)
}
