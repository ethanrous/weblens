package media

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ethrousseau/weblens/api"
	"github.com/ethrousseau/weblens/api/internal"
	error2 "github.com/ethrousseau/weblens/api/internal/werror"
	"github.com/ethrousseau/weblens/api/internal/wlog"
	"github.com/ethrousseau/weblens/api/types"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type VideoStreamer struct {
	media *weblens.Media
	encodingBegun bool
	streamDirPath string
	err           error
}

func NewVideoStreamer(media types.Media) *VideoStreamer {
	realM := media.(*weblens.Media)

	if realM.streamer != nil {
		return realM.streamer
	}

	vs := &VideoStreamer{
		media: realM,
	}

	realM.streamer = vs
	return vs
}

func (vs *VideoStreamer) transcodeChunks(f *fileTree.WeblensFile, speed string) {
	defer func() { vs.encodingBegun = false }()

	err := os.Mkdir(vs.streamDirPath, os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		vs.err = err
		return
	}

	autioRate := 128000
	outErr := bytes.NewBuffer(nil)
	err = ffmpeg.Input(f.GetAbsPath(), ffmpeg.KwArgs{"ss": 0}).Output(
		vs.streamDirPath+"%03d.ts", ffmpeg.KwArgs{
			"c:v":                "libx264",
			"b:v": internal.GetVideoConstBitrate(),
			"b:a":                autioRate,
			"crf":                18,
			"preset":             speed,
			"segment_list_flags": "+live",
			"format":             "segment",
			"segment_format":     "mpegts",
			"hls_init_time":      5,
			"hls_time":           5,
			"hls_list_size":      0,
			"segment_list":       filepath.Join(vs.streamDirPath, "list.m3u8"),
		},
	).WithErrorOutput(outErr).Run()

	if err != nil {
		wlog.Error.Println(outErr.String())
		vs.err = err
	}
}

func (vs *VideoStreamer) Encode() *VideoStreamer {
	if vs.streamDirPath == "" && !vs.encodingBegun {
		f := types.SERV.FileTree.Get(vs.media.FileIds[0])
		vs.streamDirPath = fmt.Sprintf("%s/%s-stream/", internal.GetCacheDir(), vs.media.ID())
		vs.encodingBegun = true
		go vs.transcodeChunks(f, "ultrafast")
		// time.Sleep(time.Millisecond * 100)
		// go vs.transcodeChunks(f, "veryslow")
	}

	return vs
}

func (vs *VideoStreamer) GetEncodeDir() string {
	return vs.streamDirPath
}

func (vs *VideoStreamer) Err() error {
	return vs.err
}

func (vs *VideoStreamer) IsTranscoding() bool {
	return vs.encodingBegun
}

func (vs *VideoStreamer) probeSourceBitrate(f *fileTree.WeblensFile) (int, error) {
	probeJson, err := ffmpeg.Probe(f.GetAbsPath())
	if err != nil {
		return 0, err
	}
	probeResult := map[string]any{}
	err = json.Unmarshal([]byte(probeJson), &probeResult)
	if err != nil {
		return 0, err
	}

	formatChunk, ok := probeResult["format"].(map[string]any)
	if !ok {
		return 0, error2.NewWeblensError("invalid movie format")
	}
	bitRate, err := strconv.ParseInt(formatChunk["bit_rate"].(string), 10, 64)
	if err != nil {
		return 0, err
	}
	return int(bitRate), nil
}
