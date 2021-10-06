// +build !test

package pipeline

import (
	"github.com/tinyzimmer/go-gst/gst"
)

type Output struct {
	mux      *gst.Element
	sink     *gst.Element
	audioPad string
	videoPad string
}

func (s *Output) GetAudioSinkPad() *gst.Pad {
	return s.mux.GetRequestPad(s.audioPad)
}

func (s *Output) GetVideoSinkPad() *gst.Pad {
	return s.mux.GetRequestPad(s.videoPad)
}

// TODO: multiple rtmp
func getRtmpOutput(rtmp []string) (*Output, error) {
	mux, err := gst.NewElement("flvmux")
	if err != nil {
		return nil, err
	}
	err = mux.Set("streamable", true)
	if err != nil {
		return nil, err
	}

	sink, err := gst.NewElement("rtmpsink")
	if err != nil {
		return nil, err
	}
	err = sink.Set("location", rtmp[0])
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

func getFileOutput(filename string) (*Output, error) {
	mux, err := gst.NewElement("mp4mux")
	if err != nil {
		return nil, err
	}
	err = mux.SetProperty("faststart", true)
	if err != nil {
		return nil, err
	}

	sink, err := gst.NewElement("filesink")
	if err != nil {
		return nil, err
	}
	err = sink.SetProperty("location", filename)
	if err != nil {
		return nil, err
	}
	err = sink.SetProperty("sync", false)
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
