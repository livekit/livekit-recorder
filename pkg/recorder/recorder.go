package recorder

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/logger"
	livekit "github.com/livekit/protocol/proto"
	"github.com/livekit/protocol/utils"
	"github.com/tinyzimmer/go-gst/gst"

	"github.com/livekit/livekit-recorder/pkg/config"
)

type Recorder struct {
	conf *config.Config

	xvfb         *exec.Cmd
	chromeCancel func()
	pipeline     *gst.Pipeline
}

func NewRecorder(conf *config.Config) *Recorder {
	return &Recorder{
		conf: conf,
	}
}

// Run blocks until the recording is complete, then cleans up chrome and xvfb
func (r *Recorder) Init(req *livekit.StartRecordingRequest) error {
	config.UpdateRequestParams(r.conf, req)
	width := int(req.Options.InputWidth)
	height := int(req.Options.InputHeight)

	// Xvfb
	xvfb, err := r.launchXvfb(width, height, int(req.Options.Depth))
	if err != nil {
		logger.Errorw("error launching xvfb", err)
	}
	r.xvfb = xvfb

	// Chrome
	input, err := r.getInputUrl(req)
	if err != nil {
		return err
	}
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
	err := r.runGStreamer(req)

	res := &livekit.RecordingResult{Id: recordingId}
	if err != nil {
		logger.Errorw("error launching gstreamer", err)
		res.Error = err.Error()
	} else {
		logger.Infow("recording complete")
		res.Duration = time.Since(start).Milliseconds() / 1000
	}

	if s3, ok := req.Output.(*livekit.StartRecordingRequest_S3Url); ok {
		// TODO: upload
		res.DownloadUrl = s3.S3Url
	}

	return res
}

func (r *Recorder) AddOutput(rtmp string) error {
	return nil
}

func (r *Recorder) RemoveOutput(rtmp string) error {
	return nil
}

func (r *Recorder) Stop() {
	logger.Debugw("sending EOS to pipeline")
	if p := r.pipeline; p != nil {
		p.SendEvent(gst.NewEOSEvent())
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

func (r *Recorder) getInputUrl(req *livekit.StartRecordingRequest) (string, error) {
	switch req.Input.(type) {
	case *livekit.StartRecordingRequest_Url:
		return req.Input.(*livekit.StartRecordingRequest_Url).Url, nil
	case *livekit.StartRecordingRequest_Template:
		template := req.Input.(*livekit.StartRecordingRequest_Template).Template

		var token string
		switch template.Room.(type) {
		case *livekit.RecordingTemplate_RoomName:
			var err error
			token, err = r.buildToken(template.Room.(*livekit.RecordingTemplate_RoomName).RoomName)
			if err != nil {
				return "", err
			}
		case *livekit.RecordingTemplate_Token:
			token = template.Room.(*livekit.RecordingTemplate_Token).Token
		default:
			return "", errors.New("token or room name required")
		}

		return fmt.Sprintf("https://recorder.livekit.io/#/%s?url=%s&token=%s",
			template.Layout, url.QueryEscape(r.conf.WsUrl), token), nil
	default:
		return "", errors.New("input url or template required")
	}
}

func (r *Recorder) buildToken(roomName string) (string, error) {
	f := false
	t := true
	grant := &auth.VideoGrant{
		RoomRecord:   true,
		Room:         roomName,
		CanPublish:   &f,
		CanSubscribe: &t,
		Hidden:       true,
	}

	at := auth.NewAccessToken(r.conf.ApiKey, r.conf.ApiSecret).
		AddGrant(grant).
		SetIdentity(utils.NewGuid(utils.RecordingPrefix)).
		SetValidFor(24 * time.Hour)

	return at.ToJWT()
}
