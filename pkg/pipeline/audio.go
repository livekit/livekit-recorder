// +build !test

package pipeline

import (
	"fmt"

	"github.com/tinyzimmer/go-gst/gst"
)

type AudioSource struct {
	elements   []*gst.Element
	srcElement *gst.Element
}

func (s *AudioSource) LinkElements() error {
	return gst.ElementLinkMany(s.elements...)
}

func (s *AudioSource) GetSourcePad() *gst.Pad {
	return s.srcElement.GetStaticPad("src")
}

func getAudioSource(bitrate, frequency int32) (*AudioSource, error) {
	pulseSrc, err := gst.NewElement("pulsesrc")
	if err != nil {
		return nil, err
	}

	audioConvert, err := gst.NewElement("audioconvert")
	if err != nil {
		return nil, err
	}

	capsFilter, err := gst.NewElement("capsfilter")
	if err != nil {
		return nil, err
	}
	capsString := fmt.Sprintf("audio/x-raw,format=S16LE,layout=interleaved,rate=%d,channels=2", frequency)
	err = capsFilter.SetProperty("caps", gst.NewCapsFromString(capsString))
	if err != nil {
		return nil, err
	}

	faac, err := gst.NewElement("faac")
	if err != nil {
		return nil, err
	}
	err = faac.SetProperty("bitrate", int(bitrate*1000))
	if err != nil {
		return nil, err
	}

	audioQueue, err := gst.NewElement("queue")
	if err != nil {
		return nil, err
	}

	return &AudioSource{
		elements:   []*gst.Element{pulseSrc, audioConvert, capsFilter, faac, audioQueue},
		srcElement: audioQueue,
	}, nil
}
