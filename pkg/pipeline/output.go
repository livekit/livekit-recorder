package pipeline

import (
	"strings"

	"github.com/tinyzimmer/go-gst/gst"
)

type Output struct {
	mux      *gst.Element
	sink     *gst.Element
	audioPad string
	videoPad string
}

func (s *Output) LinkElements() error {
	return s.mux.Link(s.sink)
}

func (s *Output) GetAudioSinkPad() *gst.Pad {
	return s.mux.GetRequestPad(s.audioPad)
}

func (s *Output) GetVideoSinkPad() *gst.Pad {
	return s.mux.GetRequestPad(s.videoPad)
}

func getOutput(location string) (*Output, error) {
	if strings.HasPrefix(location, "rtmp") {
		return rtmpSink(location)
	}
	return fileSink(location)
}

func fileSink(location string) (*Output, error) {
	mux, err := gst.NewElement("mp4mux")
	if err != nil {
		return nil, err
	}

	sink, err := gst.NewElement("filesink")
	if err != nil {
		return nil, err
	}
	err = sink.Set("location", location)
	if err != nil {
		return nil, err
	}

	return &Output{
		mux:      mux,
		sink:     sink,
		audioPad: "audio_%u",
		videoPad: "video_%u",
	}, nil
}

func rtmpSink(location string) (*Output, error) {
	mux, err := gst.NewElement("flvmux")
	if err != nil {
		return nil, err
	}

	sink, err := gst.NewElement("rtmpsink")
	if err != nil {
		return nil, err
	}
	err = sink.Set("location", location)
	if err != nil {
		return nil, err
	}

	return &Output{
		mux:      mux,
		sink:     sink,
		audioPad: "audio",
		videoPad: "video",
	}, nil
}