package pipeline

import (
	"fmt"

	"github.com/tinyzimmer/go-gst/gst"
)

// gst-launch-1.0 pulsesrc ! audioconvert ! vorbisenc ! oggmux ! filesink location=demo.ogg
func OGG() (pipeline *gst.Pipeline, err error) {
	pulsesrc, err := gst.NewElement("pulsesrc")
	if err != nil {
		return
	}

	audioconvert, err := gst.NewElement("audioconvert")
	if err != nil {
		return
	}

	vorbisenc, err := gst.NewElement("vorbisenc")
	if err != nil {
		return
	}

	oggmux, err := gst.NewElement("oggmux")
	if err != nil {
		return
	}

	filesink, err := gst.NewElementWithName("filesink", "sink")
	if err != nil {
		return
	}
	err = filesink.Set("location", "/out/demo.ogg")
	if err != nil {
		return
	}

	// build pipeline
	pipeline, err = gst.NewPipeline("pipeline")
	if err != nil {
		return
	}
	err = pipeline.AddMany(pulsesrc, audioconvert, vorbisenc, oggmux, filesink)
	if err != nil {
		return
	}

	// link elements
	err = gst.ElementLinkMany(pulsesrc, audioconvert, vorbisenc)
	if err != nil {
		return
	}
	err = oggmux.Link(filesink)
	if err != nil {
		return
	}

	// link pads
	pad := oggmux.GetRequestPad("audio_%u")
	padLink := vorbisenc.GetStaticPad("src").Link(pad)
	if padLink != gst.PadLinkOK {
		err = fmt.Errorf("pad link: %s", padLink.String())
		return
	}

	return
}
