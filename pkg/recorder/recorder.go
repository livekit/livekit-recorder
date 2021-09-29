package recorder

import (
	"context"
	"os/exec"

	"github.com/livekit/protocol/logger"
	livekit "github.com/livekit/protocol/proto"
)

type Recorder struct {
	xvfb         *exec.Cmd
	chromeCtx    context.Context
	chromeCancel func()
}

func NewRecorder(req *livekit.StartRecordingRequest) *Recorder {
	return &Recorder{}
}

func (r *Recorder) Start() error {
	err := r.LaunchXvfb(1920, 1080, 24)
	if err != nil {
		logger.Errorw("error launching xvfb", err)
		return err
	}

	err = r.LaunchChrome("https://www.youtube.com/watch?v=m4cgLL8JaVI", 1920, 1080)
	if err != nil {
		logger.Errorw("error launching chrome", err)
		return err
	}

	err = LaunchGStreamer()
	if err != nil {
		logger.Errorw("error launching gstreamer", err)
		return err
	}
	logger.Infow("recording complete")

	return nil
}

func (r *Recorder) Stop() error {
	return nil
}
