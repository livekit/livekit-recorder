// +build !test

package pipeline

import (
	"fmt"

	"github.com/tinyzimmer/go-gst/gst"
)

type VideoSource struct {
	elements   []*gst.Element
	srcElement *gst.Element
}

func (s *VideoSource) LinkElements() error {
	return gst.ElementLinkMany(s.elements...)
}

func (s *VideoSource) GetSourcePad() *gst.Pad {
	return s.srcElement.GetStaticPad("src")
}

func getVideoSource(bitrate, framerate int32) (*VideoSource, error) {
	xImageSrc, err := gst.NewElement("ximagesrc")
	if err != nil {
		return nil, err
	}
	err = xImageSrc.Set("show-pointer", false)
	if err != nil {
		return nil, err
	}

	videoConvert, err := gst.NewElement("videoconvert")
	if err != nil {
		return nil, err
	}

	capsFilter, err := gst.NewElement("capsfilter")
	if err != nil {
		return nil, err
	}
	capsString := fmt.Sprintf("video/x-raw,framerate=%d/1", framerate)
	err = capsFilter.SetProperty("caps", gst.NewCapsFromString(capsString))
	if err != nil {
		return nil, err
	}

	x264Enc, err := gst.NewElement("x264enc")
	if err != nil {
		return nil, err
	}
	x264Enc.SetArg("bitrate", fmt.Sprint(bitrate))
	x264Enc.SetArg("speed-preset", "veryfast")
	x264Enc.SetArg("tune", "zerolatency")

	videoQueue, err := gst.NewElement("queue")
	if err != nil {
		return nil, err
	}

	return &VideoSource{
		elements:   []*gst.Element{xImageSrc, videoConvert, capsFilter, x264Enc, videoQueue},
		srcElement: videoQueue,
	}, nil
}
