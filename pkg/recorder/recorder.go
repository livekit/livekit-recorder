package recorder

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"

	"github.com/livekit/protocol/logger"
	livekit "github.com/livekit/protocol/proto"
)

type Recorder struct {
	xvfb         *exec.Cmd
	chromeCtx    context.Context
	chromeCancel func()

	apiKey    string
	apiSecret string
	wsUrl     string
}

func NewRecorder(apiKey, apiSecret, wsUrl string) *Recorder {
	return &Recorder{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		wsUrl:     wsUrl,
	}
}

func (r *Recorder) Start(req *livekit.StartRecordingRequest) error {
	width := int(req.Options.InputWidth)
	height := int(req.Options.InputHeight)
	err := r.LaunchXvfb(width, height, int(req.Options.Depth))
	if err != nil {
		logger.Errorw("error launching xvfb", err)
		return err
	}

	err = r.LaunchChrome(r.getInputUrl(req), width, height)
	if err != nil {
		logger.Errorw("error launching chrome", err)
		return err
	}

	err = RunGStreamer(r.getOutputLocation(req))
	if err != nil {
		logger.Errorw("error launching gstreamer", err)
		return err
	}
	logger.Infow("recording complete")

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
			template.Layout, url.QueryEscape(r.wsUrl), token)
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

func (r *Recorder) Stop() error {
	return nil
}
