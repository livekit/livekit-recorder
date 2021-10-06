package recorder

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/livekit/protocol/logger"
	livekit "github.com/livekit/protocol/proto"

	"github.com/livekit/livekit-recorder/pkg/config"
)

type Recorder struct {
	conf *config.Config

	isStream     bool
	filename     string
	xvfb         *exec.Cmd
	chromeCancel func()
	pipeline     *Pipeline
}

func NewRecorder(conf *config.Config) *Recorder {
	return &Recorder{
		conf: conf,
	}
}

func (r *Recorder) Init(req *livekit.StartRecordingRequest) error {
	config.UpdateRequestParams(r.conf, req)
	width := int(req.Options.Width)
	height := int(req.Options.Height)

	// validate input
	input, err := r.getInputUrl(req)
	if err != nil {
		return err
	}

	// validate output
	switch req.Output.(type) {
	case *livekit.StartRecordingRequest_S3Url:
		s3 := req.Output.(*livekit.StartRecordingRequest_S3Url).S3Url
		idx := strings.LastIndex(s3, "/")
		if idx < 6 ||
			!strings.HasPrefix(s3, "s3://") ||
			!strings.HasSuffix(s3, ".mp4") {
			return errors.New("s3 output must be s3://bucket/{path/}filename.mp4")
		}
		r.filename = s3[idx+1:]
		r.isStream = false
	case *livekit.StartRecordingRequest_Rtmp:
		r.isStream = true
	case *livekit.StartRecordingRequest_File:
		filename := req.Output.(*livekit.StartRecordingRequest_File).File
		if !strings.HasSuffix(filename, ".mp4") {
			return errors.New("file output must be {path/}filename.mp4")
		}
		r.filename = filename
		r.isStream = false
	default:
		return errors.New("missing output")
	}

	if r.conf.Test {
		return nil
	}

	// Xvfb
	xvfb, err := r.launchXvfb(width, height, int(req.Options.Depth))
	if err != nil {
		logger.Errorw("error launching xvfb", err)
	}
	r.xvfb = xvfb

	// Chrome
	cancel, err := r.launchChrome(input, width, height)
	if err != nil {
		logger.Errorw("error launching chrome", err)
		return err
	}
	r.chromeCancel = cancel
	return nil
}

// Run blocks until completion
func (r *Recorder) Run(recordingId string, req *livekit.StartRecordingRequest) *livekit.RecordingResult {
	var err error
	res := &livekit.RecordingResult{Id: recordingId}
	defer logger.Infow("recording complete", "recordingId", recordingId,
		"error", res.Error, "duration", res.Duration, "url", res.DownloadUrl)

	start := time.Now()
	err = r.runGStreamer(req)
	if err != nil {
		logger.Errorw("error launching gstreamer", err)
		res.Error = err.Error()
		return res
	}

	res.Duration = time.Since(start).Milliseconds() / 1000

	if s3, ok := req.Output.(*livekit.StartRecordingRequest_S3Url); ok && !r.conf.Test {
		if err = r.upload(s3.S3Url); err != nil {
			res.Error = err.Error()
			return res
		}

		res.DownloadUrl = s3.S3Url
	}

	return res
}

// TODO
func (r *Recorder) AddOutput(rtmp string) error {
	if !r.isStream {
		return errors.New("cannot add stream output to file recording")
	}
	return nil
}

// TODO
func (r *Recorder) RemoveOutput(rtmp string) error {
	return nil
}

func (r *Recorder) Stop() {
	if p := r.pipeline; p != nil {
		p.Close()
	}
}

func (r *Recorder) Close() error {
	if r.chromeCancel != nil {
		r.chromeCancel()
		r.chromeCancel = nil
	}
	if r.xvfb != nil {
		err := r.xvfb.Process.Signal(os.Interrupt)
		if err != nil {
			return err
		}
		r.xvfb = nil
	}
	return nil
}
