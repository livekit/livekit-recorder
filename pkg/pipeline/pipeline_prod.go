// +build !test

package pipeline

import (
	"fmt"
	"time"

	"github.com/livekit/protocol/logger"
	livekit "github.com/livekit/protocol/proto"
	"github.com/tinyzimmer/go-glib/glib"
	"github.com/tinyzimmer/go-gst/gst"
)

type Pipeline struct {
	pipeline *gst.Pipeline
	audio    *AudioSource
	video    *VideoSource
	output   *Output
}

func NewRtmpPipeline(rtmp []string, options *livekit.RecordingOptions) (*Pipeline, error) {
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
	p, err := MP4()
	if err != nil {
		return nil, err
	}
	return &Pipeline{pipeline: p}, nil
	// output, err := getFileOutput(filename)
	// if err != nil {
	// 	return nil, err
	// }
	// p, err := newPipeline(output, options)
	// if err != nil {
	// 	return nil, err
	// }
	// return p, nil
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
		audio:    audioSource,
		video:    videoSource,
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
			logger.Infow("quitting")
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

	go func() {
		time.Sleep(time.Second * 15)
		logger.Infow("sending EOS")
		p.pipeline.SendEvent(gst.NewEOSEvent())
	}()

	// Block and iterate on the main loop
	loop.Run()
	return nil
}

func (p *Pipeline) Close() {
	logger.Debugw("Sending EOS to pipeline")
	p.pipeline.SendEvent(gst.NewEOSEvent())
}
