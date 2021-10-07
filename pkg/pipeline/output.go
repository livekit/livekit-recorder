// +build !test

package pipeline

import (
	"fmt"

	"github.com/tinyzimmer/go-gst/gst"
)

type Output struct {
	elements []*gst.Element
	mux      *gst.Element
	sinks    []*gst.Element
	audioPad string
	videoPad string

	// rtmp only
	tee    *gst.Element
	queues []*gst.Element
}

func (s *Output) GetAudioSinkPad() *gst.Pad {
	return s.mux.GetRequestPad(s.audioPad)
}

func (s *Output) GetVideoSinkPad() *gst.Pad {
	return s.mux.GetRequestPad(s.videoPad)
}

func (s *Output) LinkElements() error {
	if s.tee == nil {
		return s.mux.Link(s.sinks[0])
	}

	err := s.mux.Link(s.tee)
	if err != nil {
		return err
	}

	for i := range s.sinks {
		err = s.queues[i].Link(s.sinks[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Output) LinkPads() error {
	if s.tee == nil {
		return nil
	}

	for i, q := range s.queues {
		link := s.tee.GetRequestPad(fmt.Sprintf("src_%d", i)).Link(q.GetStaticPad("sink"))
		if link != gst.PadLinkOK {
			return fmt.Errorf("pad link: %s", link.String())
		}
	}

	return nil
}

func getRtmpOutput(rtmp []string) (*Output, error) {
	flvMux, err := gst.NewElement("flvmux")
	if err != nil {
		return nil, err
	}
	err = flvMux.Set("streamable", true)
	if err != nil {
		return nil, err
	}

	tee, err := gst.NewElement("tee")
	if err != nil {
		return nil, err
	}

	elements := []*gst.Element{flvMux, tee}
	queues := make([]*gst.Element, 0, len(rtmp))
	rtmpSinks := make([]*gst.Element, 0, len(rtmp))
	for _, url := range rtmp {
		queue, err := gst.NewElement("queue")
		if err != nil {
			return nil, err
		}
		queues = append(queues, queue)
		elements = append(elements, queue)

		rtmpSink, err := gst.NewElement("rtmpsink")
		if err != nil {
			return nil, err
		}
		err = rtmpSink.Set("location", url)
		if err != nil {
			return nil, err
		}
		rtmpSinks = append(rtmpSinks, rtmpSink)
		elements = append(elements, rtmpSink)
	}

	return &Output{
		elements: elements,
		mux:      flvMux,
		queues:   queues,
		sinks:    rtmpSinks,
		audioPad: "audio",
		videoPad: "video",
	}, nil
}

func getFileOutput(filename string) (*Output, error) {
	mp4Mux, err := gst.NewElement("mp4mux")
	if err != nil {
		return nil, err
	}
	err = mp4Mux.SetProperty("faststart", true)
	if err != nil {
		return nil, err
	}

	fileSink, err := gst.NewElement("filesink")
	if err != nil {
		return nil, err
	}
	err = fileSink.SetProperty("location", filename)
	if err != nil {
		return nil, err
	}
	err = fileSink.SetProperty("sync", false)
	if err != nil {
		return nil, err
	}

	return &Output{
		elements: []*gst.Element{mp4Mux, fileSink},
		mux:      mp4Mux,
		sinks:    []*gst.Element{fileSink},
		audioPad: "audio_%u",
		videoPad: "video_%u",
	}, nil
}
