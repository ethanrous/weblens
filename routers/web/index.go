// Package web provides HTTP handlers for serving the Weblens web UI and static content.
package web

import (
	"fmt"
	"strings"

	cover_model "github.com/ethanrous/weblens/models/cover"
	media_model "github.com/ethanrous/weblens/models/media"
	share_model "github.com/ethanrous/weblens/models/share"
	"github.com/ethanrous/weblens/modules/wlerrors"
	"github.com/ethanrous/weblens/services/ctxservice"
	"github.com/rs/zerolog/log"
)

const apiBasePath = "/api/v1"

func getShare(ctx ctxservice.RequestContext, proxyAddress string) (share *share_model.FileShare, url string, isFileSharePath bool) {
	path := ctx.Req.URL.Path

	if path[0] == '/' {
		path = path[1:]
	}

	url = fmt.Sprintf("%s/%s", proxyAddress, path)

	var err error

	if strings.HasPrefix(path, "files/share/") {
		isFileSharePath := true
		path = path[len("files/share/"):]
		slashIndex := strings.Index(path, "/")

		if slashIndex != -1 {
			path = path[:slashIndex]
		}

		shareID := share_model.IDFromString(path)

		share, err = share_model.GetShareByID(ctx, shareID)
		if err != nil && wlerrors.Is(err, share_model.ErrShareNotFound) {
			log.Error().Stack().Err(err).Msg("")

			return nil, url, isFileSharePath
		}
	}

	return share, url, isFileSharePath
}

func handleMediaPage(ctx ctxservice.RequestContext, mediaID string, fields *indexFields) {
	m, err := media_model.GetMediaByContentID(ctx, mediaID)
	if err != nil {
		log.Error().Stack().Err(err).Msg("")

		return
	}

	if m == nil {
		return
	}

	fields.Description = "Weblens Media"
	imgURL := fmt.Sprintf("%s/media/%s.webp?quality=thumbnail", apiBasePath, m.ContentID)

	if ctx.Query("shareID") != "" {
		imgURL += "?shareID=" + ctx.Query("shareID")
		fields.Description += " Share"
	}

	fields.Image = imgURL
}

func getIndexFields(ctx ctxservice.RequestContext, proxyAddress string) (fields indexFields) {
	path := ctx.Req.URL.Path
	share, url, isFileSharePath := getShare(ctx, proxyAddress)
	fields.URL = url

	if isFileSharePath {
		if share != nil {
			f, err := ctx.FileService.GetFileByID(ctx, share.FileID)
			if err != nil {
				log.Error().Stack().Err(err).Msg("")

				return fields
			}

			// TODO: consider sending file name in private shares. An option, perhaps?
			fields.Title = f.GetPortablePath().Filename() + " - Weblens"

			shareType := "file"
			if f.IsDir() {
				shareType = "folder"
			}

			if !share.IsPublic() {
				fields.Description = fmt.Sprintf("Weblens %s share. Sign in to view", shareType)
				fields.Image = "/static/favicon_48x48.png"

				return fields
			}

			fields.Description = fmt.Sprintf("Weblens %s share", shareType)

			var m *media_model.Media

			if f.IsDir() {
				cover, err := cover_model.GetCoverByFolderID(ctx, f.ID())
				if err == nil {
					m, _ = media_model.GetMediaByContentID(ctx, cover.CoverPhotoID)
				} else {
					fields.Image = "/static/folder.png"
				}
			} else {
				m, _ = media_model.GetMediaByContentID(ctx, f.GetContentID())
			}

			if m != nil && fields.Image == "" {
				fields.Image = fmt.Sprintf("%s/media/%s.webp?quality=thumbnail&shareID=%s", apiBasePath, m.ContentID, share.ID().Hex())
			}
		}
	} else if strings.HasPrefix(path, "media/") {
		handleMediaPage(ctx, path[len("/media/"):], &fields)
	}

	if fields.Image == "" {
		fields.Image = "/static/favicon_48x48.png"
	}

	return fields
}
