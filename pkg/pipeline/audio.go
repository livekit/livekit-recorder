package pipeline

import "github.com/tinyzimmer/go-gst/gst"

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

// TODO: bitrate and frequency
func getAudioSource() (*AudioSource, error) {
	pulseSrc, err := gst.NewElement("pulsesrc")
	if err != nil {
		return nil, err
	}

	audioConvert, err := gst.NewElement("audioconvert")
	if err != nil {
		return nil, err
	}

	faac, err := gst.NewElement("faac")
	if err != nil {
		return nil, err
	}

	audioQueue, err := gst.NewElement("queue")
	if err != nil {
		return nil, err
	}

	return &AudioSource{
		elements:   []*gst.Element{pulseSrc, audioConvert, faac, audioQueue},
		srcElement: audioQueue,
	}, nil
}
