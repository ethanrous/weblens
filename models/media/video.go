package media

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	file_model "github.com/ethanrous/weblens/models/file"
	"github.com/ethanrous/weblens/modules/errors"
	"github.com/ethanrous/weblens/modules/fs"
	"github.com/rs/zerolog"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type VideoStreamer struct {
	err           error
	file          *file_model.WeblensFileImpl
	streamDirPath fs.Filepath
	listFileCache []byte
	updateMu      sync.RWMutex
	encodingBegun atomic.Bool

	log zerolog.Logger
}

func NewVideoStreamer(file *file_model.WeblensFileImpl, thumbsPath fs.Filepath) *VideoStreamer {
	streamDir := thumbsPath.Child(file.GetContentId()+"-stream", true)

	return &VideoStreamer{
		file:          file,
		streamDirPath: streamDir,
	}
}

func (vs *VideoStreamer) transcodeChunks(f *file_model.WeblensFileImpl, speed string) {
	defer func() {
		vs.encodingBegun.Store(false)

		e := recover()
		if e == nil {
			return
		}

		err, ok := e.(error)
		if !ok {
			vs.log.Error().Msgf("transcodeChunks panicked and got non-error error: %v", e)

			return
		}

		vs.log.Error().Stack().Err(err).Msg("")
	}()

	vs.log.Debug().Func(func(e *zerolog.Event) {
		e.Msgf("Transcoding video %s => %s", f.GetPortablePath().ToAbsolute(), vs.streamDirPath)
	})

	err := os.Mkdir(vs.streamDirPath.ToAbsolute(), os.ModePerm)
	if err != nil && !os.IsExist(err) {
		vs.updateMu.Lock()
		vs.err = err
		vs.updateMu.Unlock()

		return
	}

	videoBitrate, audioBitrate, err := vs.probeSourceBitrate(f)
	if err != nil {
		vs.updateMu.Lock()
		vs.err = err
		vs.updateMu.Unlock()

		return
	}

	vs.log.Trace().Func(func(e *zerolog.Event) { e.Msgf("Bitrate: %d %d", videoBitrate, audioBitrate) })
	outputArgs := ffmpeg.KwArgs{
		"c:v":                "libx264",
		"b:v":                int(videoBitrate),
		"b:a":                320_000,
		"c:a":                "aac",
		"segment_list_flags": "+live",
		"format":             "segment",
		"segment_format":     "mpegts",
		"hls_init_time":      5,
		"hls_time":           5,
		"hls_list_size":      0,
		"segment_list":       filepath.Join(vs.streamDirPath.ToAbsolute(), "list.m3u8"),
		"crf":                18,
		"preset":             speed,
	}

	outErr := bytes.NewBuffer(nil)
	err = ffmpeg.Input(f.GetPortablePath().ToAbsolute(), ffmpeg.KwArgs{"ss": 0}).Output(vs.streamDirPath.ToAbsolute()+"%03d.ts", outputArgs).WithErrorOutput(outErr).Run()

	if err != nil {
		vs.log.Error().Msg(outErr.String())
		vs.updateMu.Lock()
		vs.err = err
		vs.updateMu.Unlock()
	}
}

func (vs *VideoStreamer) Encode(f *file_model.WeblensFileImpl) *VideoStreamer {
	if !vs.encodingBegun.Load() {
		vs.encodingBegun.Store(true)
		go vs.transcodeChunks(f, "ultrafast")
	}

	return vs
}

func (vs *VideoStreamer) GetEncodeDir() fs.Filepath {
	return vs.streamDirPath
}

func (vs *VideoStreamer) GetChunk(chunkName string) (*os.File, error) {
	chunkPath := vs.GetEncodeDir().Child(chunkName, false)
	if _, err := os.Stat(chunkPath.ToAbsolute()); err != nil {
		vs.Encode(vs.file)

		for vs.IsTranscoding() {
			if _, err := os.Stat(chunkPath.ToAbsolute()); err == nil {
				break
			}

			if vs.Err() != nil {
				return nil, vs.Err()
			}

			time.Sleep(time.Second)
		}
	}

	return os.Open(chunkPath.ToAbsolute())
}

func (vs *VideoStreamer) GetListFile() ([]byte, error) {
	if vs.listFileCache != nil {
		return vs.listFileCache, nil
	}

	listPath := vs.GetEncodeDir().Child("list.m3u8", false)
	if _, err := os.Stat(listPath.ToAbsolute()); err != nil {
		vs.Encode(vs.file)

		for vs.IsTranscoding() {
			if _, err := os.Stat(listPath.ToAbsolute()); err == nil {
				break
			}

			if vs.Err() != nil {
				return nil, vs.Err()
			}

			time.Sleep(time.Second)
		}
	}

	listFile, err := os.ReadFile(listPath.ToAbsolute())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Cache the list file only if transcoding is done
	// if bytes.HasSuffix(listFile, []byte("ENDLIST")) {
	// 	vs.listFileCache = listFile
	// }

	// Cache the list file only if transcoding is finished and no errors
	if !vs.IsTranscoding() && vs.Err() == nil {
		vs.listFileCache = listFile
	}

	return listFile, nil
}

func (vs *VideoStreamer) Err() error {
	vs.updateMu.RLock()
	defer vs.updateMu.RUnlock()
	return vs.err
}

func (vs *VideoStreamer) IsTranscoding() bool {
	return vs.encodingBegun.Load()
}

func (vs *VideoStreamer) probeSourceBitrate(f *file_model.WeblensFileImpl) (videoBitrate int64, audioBitrate int64, err error) {
	vs.log.Debug().Func(func(e *zerolog.Event) { e.Msgf("Probing %s", f.GetPortablePath().ToAbsolute()) })

	probeJson, err := ffmpeg.Probe(f.GetPortablePath().ToAbsolute())
	if err != nil {
		return 0, 0, err
	}

	probeResult := map[string]any{}

	err = json.Unmarshal([]byte(probeJson), &probeResult)
	if err != nil {
		return 0, 0, err
	}

	formatChunk, ok := probeResult["format"].(map[string]any)
	if !ok {
		return 0, 0, errors.New("invalid movie format")
	}

	streamsChunk, ok := probeResult["streams"].([]any)
	if !ok {
		return 0, 0, errors.New("invalid movie format")
	}

	bitRateStr, ok := formatChunk["bit_rate"].(string)
	if !ok {
		return 0, 0, errors.New("bitrate does not exist or is not a string")
	}

	videoBitrate, err = strconv.ParseInt(bitRateStr, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	audioBitrate = 320_000

	for _, stream := range streamsChunk {
		streamMap := stream.(map[string]any)
		if streamMap["codec_type"].(string) == "audio" {
			bitRate, ok := streamMap["bit_rate"].(string)
			if !ok {
				continue
			}

			audioBitrate, err = strconv.ParseInt(bitRate, 10, 64)
			if err != nil {
				return 0, 0, err
			}

			break
		}
	}

	return videoBitrate, audioBitrate, nil
}
