// +build !test

package pipeline

import (
	"fmt"

	"github.com/tinyzimmer/go-gst/gst"
)

type OutputBin struct {
	isStream bool
	bin      *gst.Bin

	// file only
	fileSink *gst.Element

	// rtmp only
	tee       *gst.Element
	queues    []*gst.Element
	rtmpSinks []*gst.Element
}

func newFileOutputBin(filename string) (*OutputBin, error) {
	// create elements
	sink, err := gst.NewElement("filesink")
	if err != nil {
		return nil, err
	}
	if err = sink.SetProperty("location", filename); err != nil {
		return nil, err
	}
	if err = sink.SetProperty("sync", false); err != nil {
		return nil, err
	}

	// create bin
	bin := gst.NewBin("output")
	if err = bin.Add(sink); err != nil {
		return nil, err
	}

	// add ghost pad
	ghostPad := gst.NewGhostPad("sink", sink.GetStaticPad("sink"))
	if !bin.AddPad(ghostPad.Pad) {
		return nil, ErrGhostPadFailed
	}

	return &OutputBin{
		isStream: false,
		fileSink: sink,
		bin:      bin,
	}, nil
}

func newRtmpOutputBin(urls []string) (*OutputBin, error) {
	// create elements
	tee, err := gst.NewElement("tee")
	if err != nil {
		return nil, err
	}

	queues := make([]*gst.Element, 0, len(urls))
	sinks := make([]*gst.Element, 0, len(urls))
	for _, url := range urls {
		queue, err := gst.NewElement("queue")
		if err != nil {
			return nil, err
		}
		queues = append(queues, queue)

		sink, err := gst.NewElement("rtmpsink")
		if err != nil {
			return nil, err
		}
		err = sink.Set("location", url)
		if err != nil {
			return nil, err
		}
		sinks = append(sinks, sink)
	}

	// create bin
	bin := gst.NewBin("output")
	if err = bin.Add(tee); err != nil {
		return nil, err
	}
	if err = bin.AddMany(queues...); err != nil {
		return nil, err
	}
	if err = bin.AddMany(sinks...); err != nil {
		return nil, err
	}

	// add ghost pad
	ghostPad := gst.NewGhostPad("sink", tee.GetStaticPad("sink"))
	if !bin.AddPad(ghostPad.Pad) {
		return nil, ErrGhostPadFailed
	}

	return &OutputBin{
		isStream:  true,
		tee:       tee,
		queues:    queues,
		rtmpSinks: sinks,
		bin:       bin,
	}, nil
}

func (b *OutputBin) Link() error {
	if !b.isStream {
		return nil
	}

	for i, q := range b.queues {
		// link queue to rtmp sink
		if err := q.Link(b.rtmpSinks[i]); err != nil {
			return err
		}

		// link tee to queue
		link := b.tee.GetRequestPad(fmt.Sprintf("src_%d", i)).Link(q.GetStaticPad("sink"))
		if link != gst.PadLinkOK {
			return fmt.Errorf("pad link: %s", link.String())
		}
	}

	return nil
}

func (b *OutputBin) AddRtmpSink(url string) error {
	if !b.isStream {
		return ErrCannotAddToFile
	}
	return nil
}

func (b *OutputBin) RemoveRtmpSink(url string) error {
	if !b.isStream {
		return ErrCannotRemoveFromFile
	}
	return nil
}
