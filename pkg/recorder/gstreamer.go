package recorder

import (
	"fmt"
	"os"
	"time"

	"github.com/livekit/protocol/logger"
	"github.com/tinyzimmer/go-glib/glib"
	"github.com/tinyzimmer/go-gst/gst"

	"github.com/livekit/livekit-recorder/pkg/recorder/pipelines"
)

func LaunchGStreamer() error {
	logger.Debugw("launching gstreamer")
	_ = os.Setenv("GST_DEBUG", "3")

	gst.Init(nil)
	pipeline, err := pipelines.MP4()
	if err != nil {
		return err
	}

	// run
	loop := glib.NewMainLoop(glib.MainContextDefault(), false)
	pipeline.GetPipelineBus().AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageEOS:
			logger.Infow("EOS received")
			_ = pipeline.BlockSetState(gst.StateNull)
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

	// Start the pipeline
	err = pipeline.SetState(gst.StatePlaying)
	if err != nil {
		return err
	}

	go func() {
		time.Sleep(time.Second * 15)
		logger.Infow("sending EOS")
		pipeline.SendEvent(gst.NewEOSEvent())
	}()

	// Block and iterate on the main loop
	loop.Run()
	return nil
}
