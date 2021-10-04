package recorder

import (
	"fmt"

	"github.com/livekit/protocol/logger"
	livekit "github.com/livekit/protocol/proto"
	"github.com/tinyzimmer/go-glib/glib"
	"github.com/tinyzimmer/go-gst/gst"

	"github.com/livekit/livekit-recorder/pkg/pipeline"
)

// TODO: error handling, GPU support
func (r *Recorder) runGStreamer(req *livekit.StartRecordingRequest) error {
	logger.Debugw("launching gstreamer")

	// build pipeline
	gst.Init(nil)
	p, err := pipeline.NewPipeline("TODO")
	if err != nil {
		return err
	}
	r.pipeline = p

	// message watch
	loop := glib.NewMainLoop(glib.MainContextDefault(), false)
	p.GetPipelineBus().AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageEOS:
			logger.Infow("EOS received")
			_ = p.BlockSetState(gst.StateNull)
			logger.Infow("pipeline stopped")
			loop.Quit()
		case gst.MessageError:
			gErr := msg.ParseError()
			logger.Errorw("message error", gErr, "debug", gErr.DebugString())
			loop.Quit()
		default:
			fmt.Println(msg)
		}
		return true
	})

	// start playing
	err = p.SetState(gst.StatePlaying)
	if err != nil {
		return err
	}

	// Block and iterate on the main loop
	loop.Run()
	return nil
}
