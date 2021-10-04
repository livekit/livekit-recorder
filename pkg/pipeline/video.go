package pipeline

import "github.com/tinyzimmer/go-gst/gst"

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

// TODO: scaling
func getVideoSource() (*VideoSource, error) {
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

	x264Enc, err := gst.NewElement("x264enc")
	if err != nil {
		return nil, err
	}
	x264Enc.SetArg("speed-preset", "veryfast")
	x264Enc.SetArg("tune", "zerolatency")

	videoQueue, err := gst.NewElement("queue")
	if err != nil {
		return nil, err
	}

	return &VideoSource{
		elements:   []*gst.Element{xImageSrc, videoConvert, x264Enc, videoQueue},
		srcElement: videoQueue,
	}, nil
}
