package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	file_model "github.com/ethanrous/weblens/models/file"
	media_model "github.com/ethanrous/weblens/models/media"
	"github.com/ethanrous/weblens/modules/errors"
	context_service "github.com/ethanrous/weblens/services/context"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// StreamVideo returns the video streamer object for the given media.
func StreamVideo(ctx context.Context, m *media_model.Media) (*media_model.VideoStreamer, error) {
	appCtx, ok := context_service.FromContext(ctx)
	if !ok {
		return nil, errors.WithStack(context_service.ErrNoContext)
	}

	if !media_model.ParseMime(m.MimeType).IsVideo {
		return nil, errors.WithStack(ErrMediaNotVideo)
	}

	cache := appCtx.GetCache(videoStreamerContextKey)

	streamerAny, ok := cache.Get(m.ID())
	if ok {
		return streamerAny.(*media_model.VideoStreamer), nil
	}

	f, err := appCtx.FileService.GetFileByID(ctx, m.FileIDs[0])
	if err != nil {
		return nil, err
	}

	streamer := media_model.NewVideoStreamer(f, file_model.ThumbsDirPath)

	cache.Set(m.ID(), streamer)

	return streamer, nil
}

func generateVideoThumbnail(filepath string) ([]byte, error) {
	const frameNum = 10

	buf := bytes.NewBuffer(nil)
	errOut := bytes.NewBuffer(nil)

	// Get the 10th frame of the video and save it to the cache as the thumbnail
	// "Highres" for video is the video itself
	err := ffmpeg.Input(filepath).Filter(
		"select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", frameNum)},
	).Output(
		"pipe:", ffmpeg.KwArgs{"frames:v": 1, "format": "image2", "vcodec": "mjpeg"},
	).WithOutput(buf).WithErrorOutput(errOut).Run()
	if err != nil {
		return nil, errors.WithStack(errors.New(err.Error() + errOut.String()))
	}

	return buf.Bytes(), nil
}

func getVideoDurationMs(filepath string) (int, error) {
	probeJSON, err := ffmpeg.Probe(filepath)
	if err != nil {
		return 0, err
	}

	probeResult := map[string]any{}

	err = json.Unmarshal([]byte(probeJSON), &probeResult)
	if err != nil {
		return 0, err
	}

	formatChunk, ok := probeResult["format"].(map[string]any)
	if !ok {
		return 0, errors.Errorf("invalid movie format")
	}

	duration, err := strconv.ParseFloat(formatChunk["duration"].(string), 32)
	if err != nil {
		return 0, err
	}

	return int(duration) * 1000, nil
}
