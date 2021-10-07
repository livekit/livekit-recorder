package pipeline

import (
	"fmt"

	"github.com/tinyzimmer/go-gst/gst"
)

func MP4() (pipeline *gst.Pipeline, err error) {
	// audio elements
	pulsesrc, err := gst.NewElement("pulsesrc")
	if err != nil {
		return
	}

	audioconvert, err := gst.NewElement("audioconvert")
	if err != nil {
		return
	}

	faac, err := gst.NewElement("faac")
	if err != nil {
		return
	}

	audioqueue, err := gst.NewElement("queue")
	if err != nil {
		return
	}

	// video elements
	ximagesrc, err := gst.NewElement("ximagesrc")
	if err != nil {
		return
	}
	err = ximagesrc.Set("show-pointer", false)
	if err != nil {
		return
	}

	videoconvert, err := gst.NewElement("videoconvert")
	if err != nil {
		return
	}

	x264enc, err := gst.NewElement("x264enc")
	if err != nil {
		return
	}
	x264enc.SetArg("speed-preset", "veryfast")
	x264enc.SetArg("tune", "zerolatency")

	videoqueue, err := gst.NewElement("queue")
	if err != nil {
		return
	}

	mp4mux, err := gst.NewElement("mp4mux")
	if err != nil {
		return
	}

	filesink, err := gst.NewElement("filesink")
	if err != nil {
		return
	}
	err = filesink.Set("location", "/out/demo.mp4")
	if err != nil {
		return
	}

	// build pipeline
	pipeline, err = gst.NewPipeline("pipeline")
	if err != nil {
		return
	}
	err = pipeline.AddMany(
		ximagesrc, videoconvert, x264enc, videoqueue,
		pulsesrc, audioconvert, faac, audioqueue,
		mp4mux, filesink,
	)
	if err != nil {
		return
	}

	// link elements
	err = gst.ElementLinkMany(ximagesrc, videoconvert, x264enc, videoqueue)
	if err != nil {
		return
	}
	err = gst.ElementLinkMany(pulsesrc, audioconvert, faac, audioqueue)
	if err != nil {
		return
	}
	err = mp4mux.Link(filesink)
	if err != nil {
		return
	}

	// link pads
	if link := audioqueue.GetStaticPad("src").Link(mp4mux.GetRequestPad("audio_%u")); link != gst.PadLinkOK {
		err = fmt.Errorf("pad link: %s", link.String())
		return
	}
	if link := videoqueue.GetStaticPad("src").Link(mp4mux.GetRequestPad("video_%u")); link != gst.PadLinkOK {
		err = fmt.Errorf("pad link: %s", link.String())
		return
	}

	return
}
