package jobs

import (
	"slices"

	"github.com/ethanrous/weblens/models/embedding"
	"github.com/ethanrous/weblens/models/job"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/models/task"
	"github.com/ethanrous/weblens/modules/websocket"
	"github.com/ethanrous/weblens/modules/wlerrors"
	context_service "github.com/ethanrous/weblens/services/ctxservice"
	media_service "github.com/ethanrous/weblens/services/media"
	"github.com/ethanrous/weblens/services/notify"
	"github.com/ethanrous/weblens/services/reshape"
)

// IndexFile scans an individual file and processes its metadata.
func IndexFile(tsk *task.Task) {
	reportSubscanStatus(tsk)

	meta := tsk.GetMeta().(job.IndexMeta)

	ctx, ok := context_service.FromContext(tsk.Ctx)
	if !ok {
		tsk.Fail(wlerrors.New("failed to get context"))

		return
	}

	tsk.SetResult(task.Result{
		"filepath": meta.File.GetPortablePath().String(),
	})

	err := ScanFileTsk(ctx, meta)
	if err != nil {
		tsk.Fail(err)

		return
	}

	dispatchEmbedTask(ctx, meta.File)
	tsk.Success()
}

// ScanFileTsk is the internal implementation for scanning a file with the given context and metadata.
func ScanFileTsk(ctx context_service.AppContext, meta job.IndexMeta) error {
	if !media_model.ParseExtension(meta.File.GetPortablePath().Ext()).Displayable {
		return wlerrors.WithStack(media_model.ErrNotDisplayable)
	}

	existingMedia, err := media_model.GetMediaByContentID(ctx, meta.File.GetContentID())

	if err == nil && !meta.ForceReIndex && existingMedia.IsSufficentlyProcessed(false, false) {
		ctx.Log().Trace().Msgf("Media [%s] already sufficiently processed, skipping", existingMedia.ID())

		if !slices.Contains(existingMedia.FileIDs, meta.File.ID()) {
			err = existingMedia.AddFileToMedia(ctx, meta.File.ID())
			if err != nil {
				return err
			}
		}

		return nil
	}

	media := existingMedia
	mediaIsNew := media == nil
	isCached := false

	if mediaIsNew || meta.ForceReIndex {
		if meta.ForceReIndex && existingMedia != nil {
			// Drop the old media's stale embeddings and cache before rebuilding it fresh.
			if err := embedding.DeleteAllForSources(ctx, []string{string(existingMedia.ContentID), meta.File.ID()}); err != nil {
				return wlerrors.Errorf("failed to delete existing embeddings for re-index: %w", err)
			}

			if err := media_service.PurgeCache(ctx, existingMedia); err != nil {
				ctx.Log().Warn().Err(err).Msgf("Failed to purge cache for media %s during re-index", existingMedia.ContentID)
			}
		}

		media, err = media_service.NewMediaFromFile(ctx, meta.File)
		if err != nil {
			return err
		}
	} else {
		if !slices.Contains(existingMedia.FileIDs, meta.File.ID()) {
			err = existingMedia.AddFileToMedia(ctx, meta.File.ID())
			if err != nil {
				return err
			}
		}

		// Check if the media has thumbnails cached on the filesystem. If not, we need to regenerate them.
		isCached, err = media_service.IsCached(ctx, media)
		if err != nil {
			return err
		}
	}

	// Generate the thumbnails if they do not exist
	if !isCached {
		_, err = media_service.HandleCacheCreation(ctx, media, meta.File)
		if err != nil {
			return err
		}
	}

	err = media_model.SaveMedia(ctx, media)
	if err != nil {
		ctx.Log().Debug().Msgf("Failed to save media %s - %s", media.MediaID, media.ID())

		return wlerrors.Errorf("failed to save media: %w", err)
	}

	mediaInfo := reshape.MediaToMediaInfo(media)

	o := notify.FileNotificationOptions{MediaInfo: mediaInfo}

	fInfo, err := reshape.WeblensFileToFileInfo(ctx, meta.File)
	if err != nil {
		return err
	}

	notif := notify.NewFileNotification(ctx, fInfo, websocket.FileUpdatedEvent, o)
	ctx.Notify(ctx, notif...)

	return nil
}
