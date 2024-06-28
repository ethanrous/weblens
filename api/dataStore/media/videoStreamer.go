package media

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/ethrousseau/weblens/api/types"
	"github.com/ethrousseau/weblens/api/util"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type VideoStreamer struct {
	bytesTotal    int
	currentPos    int64
	writeHead     int64
	buf           *lockingBuf
	encoded       []byte
	media         *Media
	encodingBegun bool
	seekLock      *sync.Mutex
	bufferLock    *sync.Mutex
	modified      time.Time
	sinceLastRead time.Time
}

type lockingBuf struct {
	buf     *bytes.Buffer
	bufLock *sync.Mutex
}

func newWrappedBuf() *lockingBuf {
	return &lockingBuf{bytes.NewBuffer(nil), &sync.Mutex{}}
}

func (wb *lockingBuf) Write(p []byte) (int, error) {
	wb.bufLock.Lock()
	defer wb.bufLock.Unlock()

	wrote, err := wb.buf.Write(p)
	return wrote, err
}

func (wb *lockingBuf) Read(p []byte) (int, error) {
	wb.bufLock.Lock()
	defer wb.bufLock.Unlock()
	return wb.buf.Read(p)
}

func (wb *lockingBuf) Len() int {
	return wb.buf.Len()
}

func NewVideoStreamer(media types.Media) *VideoStreamer {
	realM := media.(*Media)

	if realM.streamer != nil {
		return realM.streamer
	}

	vs := &VideoStreamer{
		bytesTotal: -1,
		media:      realM,
		buf:        newWrappedBuf(),
		seekLock:   &sync.Mutex{},
		bufferLock: &sync.Mutex{},
	}

	realM.streamer = vs
	return vs
}

func (vs *VideoStreamer) transcodeByteRange(f types.WeblensFile) {
	outErr := bytes.NewBuffer(nil)

	autioRate := 128000
	err := ffmpeg.Input(f.GetAbsPath()).Output(
		"pipe:", ffmpeg.KwArgs{
			"c:v":          "libx264",
			"pix_fmt":      "yuv420p",
			"b:v":          util.GetVideoConstBitrate(),
			"preset":       "ultrafast",
			"movflags":     "frag_keyframe+empty_moov+use_metadata_tags",
			"map_metadata": 0,
			"b:a":          autioRate,
			"format":       "mp4",
		},
	).WithOutput(vs.buf).WithErrorOutput(outErr).Run()

	if err != nil {
		util.Error.Println(outErr.String())
		vs.encodingBegun = false
		return
	}

	vs.bufferLock.Lock()
	vs.encodingBegun = false
	// vs.modified = time.Now()
	vs.bufferLock.Unlock()

	return
}

func (vs *VideoStreamer) bufferLoader() {
	localBuf := make([]byte, 32768)
	readSoFar := 0
	for vs.encodingBegun || vs.buf.Len() != 0 {
		vs.bufferLock.Lock()
		if vs.buf.Len() < 32768 && vs.encodingBegun {
			vs.bufferLock.Unlock()
			time.Sleep(time.Millisecond * 100)
			continue
		}

		read, err := vs.buf.Read(localBuf)
		if err != nil {
			vs.bufferLock.Unlock()
			util.ErrTrace(err)
			readSoFar = -1
			break
		}

		vs.encoded = append(vs.encoded, localBuf[:read]...)
		readSoFar += read
		vs.bufferLock.Unlock()
	}
	vs.bytesTotal = readSoFar
	util.Debug.Println("Finished reading:", readSoFar)
}

func (vs *VideoStreamer) loadFromFile(f types.WeblensFile) error {
	buf, err := f.ReadAll()
	if err != nil {
		return err
	}

	vs.encoded = buf
	vs.bytesTotal = len(vs.encoded)

	return nil
}

func (vs *VideoStreamer) PreLoadBuf() int {
	vs.seekLock.Lock()
	if vs.bytesTotal == -1 && !vs.encodingBegun {
		f := types.SERV.FileTree.Get(vs.media.FileIds[0])
		// If we are within ~25% of our bitrate target, just serve the video
		if bitRate, _ := vs.probeSourceBitrate(f); int(float64(bitRate)*0.75) < util.GetVideoConstBitrate() {
			err := vs.loadFromFile(f)
			if err != nil {
				util.ShowErr(err)
			}
		} else {
			vs.modified = time.Now()
			vs.encodingBegun = true
			go vs.transcodeByteRange(f)
			go vs.bufferLoader()
		}

	}

	return vs.bytesTotal
}

func (vs *VideoStreamer) RelinquishStream() {
	vs.seekLock.Unlock()
}

func (vs *VideoStreamer) Read(buf []byte) (int, error) {
	for len(vs.encoded)-int(vs.currentPos) < len(buf) && vs.bytesTotal == -1 {
		time.Sleep(time.Millisecond * 10)
	}

	end := min(
		int(vs.currentPos)+len(buf),
		int(vs.currentPos)+(len(vs.encoded)-int(vs.currentPos)),
		len(vs.encoded),
	)

	util.Debug.Println("Time since last read", time.Now().Sub(vs.sinceLastRead))
	vs.sinceLastRead = time.Now()
	copyLen, err := bytes.NewBuffer(vs.encoded[vs.currentPos:end]).Read(buf)
	// if err != nil && err != io.EOF {
	if err != nil {
		util.ShowErr(err)
		return 0, err
	}

	err = nil

	vs.currentPos += int64(copyLen)
	if vs.bytesTotal != -1 && vs.currentPos >= int64(vs.bytesTotal) || copyLen < len(buf) {
		// util.Debug.Println("Sending EOF...")
		err = io.EOF
	}
	util.ShowErr(err)
	return copyLen, err
}

func (vs *VideoStreamer) ReadRange(start, end int) *bytes.Buffer {
	if start > end || end > len(vs.encoded) {
		util.Error.Println("Bad read range")
		return nil
	}

	return bytes.NewBuffer(vs.encoded[start:end])
}

func (vs *VideoStreamer) Seek(offset int64, whence int) (int64, error) {
	if whence == io.SeekStart {
		vs.currentPos = offset
	} else {
		return 0, fmt.Errorf("video streamer seek: invalid whence")
	}

	return vs.currentPos, nil
}

func (vs *VideoStreamer) Len() int {
	return len(vs.encoded)
}

func (vs *VideoStreamer) Modified() time.Time {
	return vs.modified
}

func (vs *VideoStreamer) IsTranscoding() bool {
	return vs.encodingBegun
}

func (vs *VideoStreamer) probeSourceBitrate(f types.WeblensFile) (int, error) {

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
		return 0, types.NewWeblensError("invalid movie format")
	}
	bitRate, err := strconv.ParseInt(formatChunk["bit_rate"].(string), 10, 64)
	if err != nil {
		return 0, err
	}
	return int(bitRate), nil
}
