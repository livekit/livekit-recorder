// +build !test

package pipeline

import (
	"errors"
	"fmt"

	"github.com/livekit/protocol/logger"
	livekit "github.com/livekit/protocol/proto"
	"github.com/tinyzimmer/go-glib/glib"
	"github.com/tinyzimmer/go-gst/gst"
)

// gst.Init needs to be called before using gst but after gst package loads
var initialized = false

type Pipeline struct {
	pipeline *gst.Pipeline
	output   *Output
	isStream bool
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
	p.isStream = true
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

	// elements must be added to pipeline before linking
	pipeline, err := gst.NewPipeline("pipeline")
	if err != nil {
		return nil, err
	}
	elements := append(audioSource.elements, videoSource.elements...)
	err = pipeline.AddMany(append(elements, output.elements...)...)
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
	err = output.LinkElements()
	if err != nil {
		return nil, err
	}

	// link mux and tee pads
	if link := audioSource.GetSourcePad().Link(output.GetAudioSinkPad()); link != gst.PadLinkOK {
		return nil, fmt.Errorf("pad link: %s", link.String())
	}
	if link := videoSource.GetSourcePad().Link(output.GetVideoSinkPad()); link != gst.PadLinkOK {
		return nil, fmt.Errorf("pad link: %s", link.String())
	}

	err = output.LinkPads()
	if err != nil {
		return nil, err
	}

	return &Pipeline{
		pipeline: pipeline,
		output:   output,
	}, nil
}

// TODO: split pipelines, restart second pipeline on failure
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

func (p *Pipeline) AddOutput(url string) error {
	if !p.isStream {
		return errors.New("AddOutput can only be called on streams")
	}

	return nil
}

func (p *Pipeline) RemoveOutput(url string) error {
	if !p.isStream {
		return errors.New("RemoveOutput can only be called on streams")
	}

	return nil
}

func (p *Pipeline) Close() {
	logger.Debugw("Sending EOS to pipeline")
	p.pipeline.SendEvent(gst.NewEOSEvent())
}
