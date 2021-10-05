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
	width := int(req.Options.InputWidth)
	height := int(req.Options.InputHeight)

	// validate input
	input, err := r.getInputUrl(req)
	if err != nil {
		return err
	}

	// validate output
	if s3, ok := req.Output.(*livekit.StartRecordingRequest_S3Url); ok {
		idx := strings.LastIndex(s3.S3Url, "/")
		if idx < 6 ||
			!strings.HasPrefix(s3.S3Url, "s3://") ||
			!strings.HasSuffix(s3.S3Url, ".mp4") {
			return errors.New("malformed s3 url, should be s3://bucket/{path/}filename.mp4")
		}
		r.filename = s3.S3Url[idx+1:]
		r.isStream = false
	} else {
		r.isStream = true
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
	start := time.Now()

	var err error
	if r.conf.Test {
		// sleep to emulate running
		time.Sleep(time.Second * 3)
	} else {
		err = r.runGStreamer(req)
	}

	res := &livekit.RecordingResult{Id: recordingId}
	if err != nil {
		logger.Errorw("error launching gstreamer", err)
		res.Error = err.Error()
	} else {
		logger.Infow("recording complete")
		res.Duration = time.Since(start).Milliseconds() / 1000
	}

	if s3, ok := req.Output.(*livekit.StartRecordingRequest_S3Url); ok && !r.conf.Test {
		if err = r.upload(s3.S3Url); err != nil {
			res.Error = err.Error()
		} else {
			res.DownloadUrl = s3.S3Url
		}
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
	logger.Debugw("sending EOS to pipeline")
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
