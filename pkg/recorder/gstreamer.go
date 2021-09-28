package recorder

import (
	"fmt"
	"time"

	"github.com/livekit/protocol/logger"
	"github.com/tinyzimmer/go-glib/glib"
	"github.com/tinyzimmer/go-gst/gst"

	"github.com/livekit/livekit-recorder/pkg/recorder/pipelines"
)

func LaunchGStreamer() error {
	logger.Debugw("launching gstreamer")

	gst.Init(nil)
	pipeline, err := pipelines.OGG()
	if err != nil {
		return err
	}

	// run
	loop := glib.NewMainLoop(glib.MainContextDefault(), false)
	pipeline.GetPipelineBus().AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageEOS:
			pipeline.BlockSetState(gst.StateNull)
			loop.Quit()
		case gst.MessageError:
			err := msg.ParseError()
			fmt.Println("ERROR:", err.Error())
			if debug := err.DebugString(); debug != "" {
				fmt.Println("DEBUG:", debug)
			}
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
		loop.Quit()
	}()

	// Block and iterate on the main loop
	loop.Run()

	return nil
}
