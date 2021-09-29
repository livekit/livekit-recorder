package pipeline

import (
	"fmt"

	"github.com/tinyzimmer/go-gst/gst"
)

func NewPipeline(location string) (*gst.Pipeline, error) {
	audioSource, err := getAudioSource()
	if err != nil {
		return nil, err
	}

	videoSource, err := getVideoSource()
	if err != nil {
		return nil, err
	}

	output, err := getOutput(location)
	if err != nil {
		return nil, err
	}

	// build pipeline
	pipeline, err := gst.NewPipeline("pipeline")
	if err != nil {
		return nil, err
	}
	elements := append(audioSource.elements, videoSource.elements...)
	err = pipeline.AddMany(append(elements, output.mux, output.sink)...)
	if err != nil {
		return nil, err
	}

	// link elements
	err = audioSource.LinkElements()
	if err != nil {
		return nil, err
	}
	err = videoSource.LinkElements()
	if err != nil {
		return nil, err
	}
	err = output.LinkElements()
	if err != nil {
		return nil, err
	}

	// link pads
	if link := audioSource.GetSourcePad().Link(output.GetAudioSinkPad()); link != gst.PadLinkOK {
		err = fmt.Errorf("pad link: %s", link.String())
		return nil, err
	}
	if link := videoSource.GetSourcePad().Link(output.GetVideoSinkPad()); link != gst.PadLinkOK {
		err = fmt.Errorf("pad link: %s", link.String())
		return nil, err
	}

	return pipeline, nil
}
