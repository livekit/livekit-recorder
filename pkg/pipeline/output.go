// +build !test

package pipeline

import (
	"fmt"

	"github.com/livekit/protocol/logger"
	"github.com/tinyzimmer/go-gst/gst"
)

type OutputBin struct {
	isStream bool
	bin      *gst.Bin

	// file only
	fileSink *gst.Element

	// rtmp only
	tee       *gst.Element
	urls      []string
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
		urls:      urls,
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
		if err := requireLink(
			b.tee.GetRequestPad(fmt.Sprintf("src_%d", i)),
			q.GetStaticPad("sink")); err != nil {
			return err
		}
	}

	return nil
}

func (b *OutputBin) AddRtmpSink(url string) error {
	if !b.isStream {
		return ErrCannotAddToFile
	}

	idx := -1
	for i, u := range b.urls {
		if u == "" && idx == -1 {
			idx = i
		}
		if u == url {
			return ErrOutputAlreadyExists
		}
	}

	queue, err := gst.NewElement("queue")
	if err != nil {
		return err
	}
	sink, err := gst.NewElement("rtmpsink")
	if err != nil {
		return err
	}
	if err = sink.Set("location", url); err != nil {
		return err
	}

	// add to bin
	if err = b.bin.AddMany(queue, sink); err != nil {
		return err
	}

	if idx == -1 {
		idx = len(b.urls)
		b.urls = append(b.urls, url)
		b.queues = append(b.queues, queue)
		b.rtmpSinks = append(b.rtmpSinks, sink)
	} else {
		b.urls[idx] = url
		b.queues[idx] = queue
		b.rtmpSinks[idx] = sink
	}

	// link queue to sink
	if err = queue.Link(sink); err != nil {
		return err
	}

	// link tee to queue
	if err = requireLink(
		b.tee.GetRequestPad(fmt.Sprintf("src_%d", idx)),
		queue.GetStaticPad("sink")); err != nil {
		return err
	}

	if err = queue.SetState(gst.StatePlaying); err != nil {
		return err
	}

	return nil
}

func (b *OutputBin) RemoveRtmpSink(url string) error {
	if !b.isStream {
		return ErrCannotRemoveFromFile
	}

	idx := -1
	for i, u := range b.urls {
		if u == url {
			idx = i
			break
		}
	}
	if idx == -1 {
		return ErrOutputNotFound
	}

	queue := b.queues[idx]
	teeSrcPad := b.tee.GetRequestPad(fmt.Sprintf("src_%d", idx))
	teeSrcPad.AddProbe(gst.PadProbeTypeBlockDownstream, b.padProbeCallback(queue))

	b.urls[idx] = ""
	b.queues[idx] = nil
	b.rtmpSinks[idx] = nil
	return nil
}

func (b *OutputBin) padProbeCallback(queue *gst.Element) gst.PadProbeCallback {
	return func(teeSrcPad *gst.Pad, info *gst.PadProbeInfo) gst.PadProbeReturn {
		// remove probe
		teeSrcPad.RemoveProbe(uint64(info.ID()))
		// add EOS probe to queue pad
		srcPad := queue.GetStaticPad("src")
		srcPad.AddProbe(gst.PadProbeTypeBlock|gst.PadProbeTypeEventDownstream, b.eventProbeCallback(queue, teeSrcPad))
		// send EOS to queue
		sinkPad := queue.GetStaticPad("sink")
		sinkPad.SendEvent(gst.NewEOSEvent())
		return gst.PadProbeOK
	}
}

func (b *OutputBin) eventProbeCallback(queue *gst.Element, teeSrcPad *gst.Pad) gst.PadProbeCallback {
	return func(queueSrcPad *gst.Pad, info *gst.PadProbeInfo) gst.PadProbeReturn {
		// skip if not EOS
		if e := info.GetEvent(); e != nil || e.Type() != gst.EventTypeEOS {
			return gst.PadProbePass
		}
		// remove probe
		queueSrcPad.RemoveProbe(uint64(info.ID()))
		// stop queue
		if err := queue.BlockSetState(gst.StateNull); err != nil {
			logger.Errorw("failed to stop rtmp queue", err)
		}
		// remove from bin
		if err := b.bin.Remove(queue); err != nil {
			logger.Errorw("failed to remove rtmp queue", err)
		}
		// release tee src pad
		b.tee.ReleaseRequestPad(teeSrcPad)
		return gst.PadProbeDrop
	}
}
