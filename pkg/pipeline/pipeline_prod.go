// +build !test

package pipeline

import (
	"fmt"

	"github.com/livekit/protocol/logger"
	livekit "github.com/livekit/protocol/proto"
	"github.com/tinyzimmer/go-glib/glib"
	"github.com/tinyzimmer/go-gst/gst"
)

var initialized = false

type Pipeline struct {
	pipeline *gst.Pipeline
	output   *Output
}

func NewRtmpPipeline(rtmp []string, options *livekit.RecordingOptions) (*Pipeline, error) {
	if !initialized {
		gst.Init(nil)
		initialized = true
	}

	output, err := getRtmpOutput(rtmp)
	if err != nil {
		return nil, err
	}
	p, err := newPipeline(output, options)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func NewFilePipeline(filename string, options *livekit.RecordingOptions) (*Pipeline, error) {
	if !initialized {
		gst.Init(nil)
		initialized = true
	}

	output, err := getFileOutput(filename)
	if err != nil {
		return nil, err
	}
	p, err := newPipeline(output, options)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func newPipeline(output *Output, options *livekit.RecordingOptions) (*Pipeline, error) {
	audioSource, err := getAudioSource(options.AudioBitrate, options.AudioFrequency)
	if err != nil {
		return nil, err
	}

	videoSource, err := getVideoSource(options.VideoBitrate, options.Framerate)
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
	err = gst.ElementLinkMany(audioSource.elements...)
	if err != nil {
		return nil, err
	}
	err = gst.ElementLinkMany(videoSource.elements...)
	if err != nil {
		return nil, err
	}
	err = output.mux.Link(output.sink)
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

	return &Pipeline{
		pipeline: pipeline,
		output:   output,
	}, nil
}

func (p *Pipeline) Start() error {
	loop := glib.NewMainLoop(glib.MainContextDefault(), false)
	p.pipeline.GetPipelineBus().AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageEOS:
			logger.Infow("EOS received")
			_ = p.pipeline.BlockSetState(gst.StateNull)
			logger.Infow("pipeline stopped")
			loop.Quit()
		case gst.MessageError:
			gErr := msg.ParseError()
			logger.Errorw("message error", gErr, "debug", gErr.DebugString())
			loop.Quit()
		default:
			fmt.Println(msg)
		}
		return true
	})

	// start playing
	err := p.pipeline.SetState(gst.StatePlaying)
	if err != nil {
		return err
	}

	// Block and iterate on the main loop
	loop.Run()
	return nil
}

func (p *Pipeline) Close() {
	logger.Debugw("Sending EOS to pipeline")
	p.pipeline.SendEvent(gst.NewEOSEvent())
}
