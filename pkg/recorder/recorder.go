package recorder

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/livekit/protocol/logger"
	livekit "github.com/livekit/protocol/proto"
	"github.com/tinyzimmer/go-gst/gst"
	"go.uber.org/atomic"

	"github.com/livekit/livekit-recorder/pkg/config"
)

type Recorder struct {
	conf *config.Config

	started  atomic.Bool
	running  atomic.Bool
	pipeline *gst.Pipeline
}

func NewRecorder(conf *config.Config) (*Recorder, error) {
	if err := os.Setenv("DISPLAY", Display); err != nil {
		return nil, err
	}

	return &Recorder{
		conf: conf,
	}, nil
}

// Run blocks until the recording is complete
func (r *Recorder) Run(req *livekit.StartRecordingRequest) error {
	if !r.started.CAS(false, true) {
		return errors.New("already started")
	}
	defer func() {
		r.pipeline = nil
		r.running.Store(false)
		r.started.Store(false)
	}()

	config.UpdateRequestParams(r.conf, req)
	width := int(req.Options.InputWidth)
	height := int(req.Options.InputHeight)

	// Xvfb
	xvfb, err := r.LaunchXvfb(width, height, int(req.Options.Depth))
	if err != nil {
		logger.Errorw("error launching xvfb", err)
		return err
	}
	defer func() {
		err := xvfb.Process.Signal(os.Interrupt)
		if err != nil {
			logger.Errorw("failed to stop xvfb", err)
		}
	}()

	// Chrome
	cancel, err := r.LaunchChrome(r.getInputUrl(req), width, height)
	if err != nil {
		logger.Errorw("error launching chrome", err)
		return err
	}
	defer cancel()

	// GStreamer
	err = r.RunGStreamer(r.getOutputLocation(req))
	if err != nil {
		logger.Errorw("error launching gstreamer", err)
		return err
	}
	logger.Infow("recording complete")
	return nil
}

func (r *Recorder) Stop() error {
	if !r.started.Load() {
		return errors.New("recorder not started")
	}

	// make sure we don't catch it between starting and running
	for !r.running.Load() {
		time.Sleep(time.Second)
	}

	logger.Debugw("sending EOS to pipeline")
	if p := r.pipeline; p != nil {
		p.SendEvent(gst.NewEOSEvent())
	}
	return nil
}

func (r *Recorder) getInputUrl(req *livekit.StartRecordingRequest) string {
	if template := req.Input.Template; template != nil {
		var token string
		if template.Token != "" {
			token = template.Token
		} else {
			token = r.buildToken(template.RoomName)
		}
		return fmt.Sprintf("https://recorder.livekit.io/#/%s?url=%s&token=%s",
			template.Layout, url.QueryEscape(r.conf.WsUrl), token)
	}
	return req.Input.Url
}

func (r *Recorder) buildToken(roomName string) string {
	return roomName // TODO
}

func (r *Recorder) getOutputLocation(req *livekit.StartRecordingRequest) string {
	if req.Output.S3Path != "" {
		return req.Output.S3Path
	}
	return req.Output.Rtmp
}
