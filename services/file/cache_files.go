package file

import (
	"context"
	"slices"

	file_model "github.com/ethanrous/weblens/models/file"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/modules/log"
	"github.com/ethanrous/weblens/modules/wlerrors"
)

// GetMediaCacheByFilename retrieves a cached media file by its thumbnail filename.
func (fs *ServiceImpl) GetMediaCacheByFilename(_ context.Context, thumbFileName string) (*file_model.WeblensFileImpl, error) {
	f := file_model.NewWeblensFile(file_model.NewFileOptions{Path: file_model.ThumbsDirPath.Child(thumbFileName, false)})
	if !f.Exists() {
		return nil, wlerrors.WithStack(file_model.ErrFileNotFound)
	}

	return f, nil
}

// NewCacheFile creates a new cache file for the specified media with the given quality and page number.
func (fs *ServiceImpl) NewCacheFile(mediaID string, quality string, pageNum int) (*file_model.WeblensFileImpl, error) {
	filename, err := media_model.FmtCacheFileName(mediaID, media_model.Quality(quality), pageNum)
	if err != nil {
		return nil, err
	}

	childPath := file_model.ThumbsDirPath.Child(filename, false)

	return touch(childPath)
}

// DeleteCacheFile removes a cache file from the filesystem.
func (fs *ServiceImpl) DeleteCacheFile(f *file_model.WeblensFileImpl) error {
	if !isCacheFile(f.GetPortablePath()) {
		return wlerrors.New("trying to delete non-cache file")
	}

	return remove(f.GetPortablePath())
}

// RemoveCacheFilesWithFilter removes cache files that match the provided content ID filter.
// If the filter is empty, no files will be removed. However, if you want to remove all cache files, pass nil as the filter.
func RemoveCacheFilesWithFilter(ctx context.Context, contentIDFilter []string) error {
	if contentIDFilter != nil && len(contentIDFilter) == 0 {
		log.FromContext(ctx).Trace().Msg("no content IDs provided, skipping cache file removal")

		return nil
	} else if contentIDFilter == nil {
		log.FromContext(ctx).Warn().Msg("no content ID filter provided, removing all cache files")
	}

	cachePaths, err := getChildFilepaths(file_model.ThumbsDirPath)
	if err != nil {
		return err
	}

	for _, cacheFilePath := range cachePaths {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		contentID, _, _, err := media_model.ParseCacheFileName(cacheFilePath.Filename())
		if err != nil {
			log.FromContext(ctx).Error().Stack().Err(err).Str("file", cacheFilePath.Filename()).Msg("failed to parse cache file name, skipping")

			continue
		}

		if contentIDFilter == nil || slices.Contains(contentIDFilter, contentID) {
			log.FromContext(ctx).Trace().Str("file", cacheFilePath.String()).Msg("removing cache file")

			err := remove(cacheFilePath)
			if err != nil {
				return err
			}
		} else {
			log.FromContext(ctx).Trace().Str("file", cacheFilePath.String()).Msg("keeping cache file")
		}
	}

	return nil
}
