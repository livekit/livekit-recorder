package recorder

import (
	"fmt"
	"time"

	"github.com/livekit/protocol/logger"
	"github.com/tinyzimmer/go-glib/glib"
	"github.com/tinyzimmer/go-gst/gst"
)

func LaunchGStreamer() error {
	logger.Debugw("launching gstreamer")

	for {
		time.Sleep(time.Minute)
		if false {
			break
		}
	}

	gst.Init(nil)

	// audio in
	audioIn, err := gst.NewElementWithName("pulsesrc", "audiosrc")
	if err != nil {
		return err
	}

	// video in
	videoIn, err := gst.NewElementWithName("ximagesrc", "videosrc")
	if err != nil {
		return err
	}
	err = videoIn.SetProperty("use-damage", false)
	if err != nil {
		return err
	}

	// mux
	mux, err := gst.NewElementWithName("mp4mux", "mux")
	if err != nil {
		return err
	}

	// output
	output, err := gst.NewElementWithName("filesink", "videosink")
	if err != nil {
		return err
	}
	err = output.SetProperty("location", "test.mp4")
	if err != nil {
		return err
	}

	// build pipeline
	pipeline, err := gst.NewPipeline("pipeline")
	if err != nil {
		return err
	}
	err = pipeline.AddMany(audioIn, videoIn, mux, output)
	if err != nil {
		return err
	}

	// link elements
	err = gst.ElementLinkMany(audioIn, mux)
	if err != nil {
		return err
	}
	err = gst.ElementLinkMany(videoIn, mux)
	if err != nil {
		return err
	}
	err = gst.ElementLinkMany(mux, output)
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

	// Block and iterate on the main loop
	loop.Run()

	time.Sleep(time.Second * 10)
	loop.Quit()

	return nil
}
