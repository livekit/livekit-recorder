package recorder

import (
	"fmt"
	"os"
	"time"

	"github.com/livekit/protocol/logger"
	"github.com/tinyzimmer/go-glib/glib"
	"github.com/tinyzimmer/go-gst/gst"

	"github.com/livekit/livekit-recorder/pkg/pipeline"
)

// TODO: bitrates, audio frequency, scaling, GPU, error handling
func RunGStreamer(location string) error {
	logger.Debugw("launching gstreamer")
	_ = os.Setenv("GST_DEBUG", "3")

	// build pipeline
	gst.Init(nil)
	p, err := pipeline.NewPipeline(location)
	if err != nil {
		return err
	}

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

	go func() {
		time.Sleep(time.Minute * 3)
		logger.Infow("sending EOS")
		p.SendEvent(gst.NewEOSEvent())
	}()

	// Block and iterate on the main loop
	loop.Run()
	return nil
}
