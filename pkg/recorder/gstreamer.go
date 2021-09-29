package recorder

import (
	"fmt"
	"os"

	"github.com/livekit/protocol/logger"
	"github.com/tinyzimmer/go-glib/glib"
	"github.com/tinyzimmer/go-gst/gst"

	"github.com/livekit/livekit-recorder/pkg/pipeline"
)

// TODO: bitrates, audio frequency, scaling, GPU, error handling
func (r *Recorder) RunGStreamer(location string) error {
	logger.Debugw("launching gstreamer")
	_ = os.Setenv("GST_DEBUG", "3")

	// build pipeline
	gst.Init(nil)
	p, err := pipeline.NewPipeline(location)
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
			logger.Infow("quitting")
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
	r.running.Store(true)
	loop.Run()
	return nil
}
