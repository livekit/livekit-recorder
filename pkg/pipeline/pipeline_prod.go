//go:build !test
// +build !test

package pipeline

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/tinyzimmer/go-gst/gst/app"

	"github.com/livekit/protocol/livekit"
	"github.com/livekit/protocol/logger"
	"github.com/tinyzimmer/go-glib/glib"
	"github.com/tinyzimmer/go-gst/gst"
)

// gst.Init needs to be called before using gst but after gst package loads
var initialized = false

type Pipeline struct {
	pipeline *gst.Pipeline
	output   *OutputBin
	input    *InputBin
	removed  map[string]bool

	started chan struct{}
	closed  chan struct{}
}

func NewRtmpPipeline(urls []string, options *livekit.RecordingOptions) (*Pipeline, error) {
	if !initialized {
		gst.Init(nil)
		initialized = true
	}

	input, err := newInputBin(true, options)
	if err != nil {
		return nil, err
	}
	output, err := newRtmpOutputBin(urls)
	if err != nil {
		return nil, err
	}

	return newPipeline(input, output)
}

func NewFilePipeline(filename string, options *livekit.RecordingOptions) (*Pipeline, error) {
	if !initialized {
		gst.Init(nil)
		initialized = true
	}

	input, err := newInputBin(false, options)
	if err != nil {
		return nil, err
	}
	output, err := newFileOutputBin(filename)
	if err != nil {
		return nil, err
	}

	return newPipeline(input, output)
}

func NewAppSinkPipeline(filename string, options *livekit.RecordingOptions) (*Pipeline, error) {
	if !initialized {
		gst.Init(nil)
		initialized = true
	}

	input, err := newInputBin(false, options)
	if err != nil {
		return nil, err
	}

	output, err := newAppSinkOutputBin(filename)
	if err != nil {
		return nil, err
	}

	return newPipeline(input, output)
}

func newPipeline(input *InputBin, output *OutputBin) (*Pipeline, error) {
	// elements must be added to pipeline before linking
	pipeline, err := gst.NewPipeline("pipeline")
	if err != nil {
		return nil, err
	}

	// add bins to pipeline
	if err = pipeline.AddMany(input.bin.Element, output.bin.Element); err != nil {
		return nil, err
	}

	// link bin elements
	if err = input.Link(); err != nil {
		return nil, err
	}
	if err = output.Link(); err != nil {
		return nil, err
	}

	// link bins
	if err = input.bin.Link(output.bin.Element); err != nil {
		return nil, err
	}

	return &Pipeline{
		pipeline: pipeline,
		output:   output,
		input:    input,
		removed:  make(map[string]bool),
		started:  make(chan struct{}, 1),
		closed:   make(chan struct{}),
	}, nil
}

func (p *Pipeline) Start() error {
	bucket := "bucket-name-here"
	object := "filename-here.mp4"
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()

	// Upload an object with storage.Writer.
	wc := client.Bucket(bucket).Object(object).NewWriter(ctx)
	wc.ContentType = "video/mp4"
	wc.ChunkSize = 0 // note retries are not supported for chunk size 0.

	asink := p.output.appSink
	queue := p.output.queue
	asink.SyncStateWithParent()
	asink.SetWaitOnEOS(false)
	queue.Link(asink.Element)
	p.input.mux.Link(queue)

	asink.SetCallbacks(&app.SinkCallbacks{
		EOSFunc: func(sink *app.Sink) {
			// Signal the pipeline that we've completed EOS.
			// (this should not be required, need to investigate)
			// pipeline.GetPipelineBus().Post(gst.NewEOSMessage(appSink))
		},
		NewSampleFunc: func(sink *app.Sink) gst.FlowReturn {
			// Pull the sample that triggered this callback
			sample := sink.PullSample()
			if sample == nil {
				return gst.FlowEOS
			}

			// Retrieve the buffer from the sample
			buffer := sample.GetBuffer()
			if buffer == nil {
				return gst.FlowError
			}

			fmt.Printf("Writing chunk\n")

			if _, err = io.Copy(wc, buffer.Reader()); err != nil {
				fmt.Errorf("io.Copy: %v", err)
			}
			fmt.Printf("%v uploaded to %v.\n", object, bucket)
			return gst.FlowOK
		},
	})

	loop := glib.NewMainLoop(glib.MainContextDefault(), false)
	p.pipeline.GetPipelineBus().AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageEOS:
			logger.Debugw("EOS received")
			_ = p.pipeline.BlockSetState(gst.StateNull)
			logger.Debugw("pipeline stopped")

			// Data can continue to be added to the file until the writer is closed.
			if err := wc.Close(); err != nil {
				fmt.Errorf("Writer.Close: %v", err)
			}

			loop.Quit()
			return false
		case gst.MessageError:
			gErr := msg.ParseError()
			handled := p.handleError(gErr)
			if !handled {
				loop.Quit()
				return false
			}
			logger.Debugw("handled error", "error", gErr.Error())
		default:
			logger.Debugw(msg.String())
		}
		return true
	})

	// start playing
	err = p.pipeline.SetState(gst.StatePlaying)
	if err != nil {
		return err
	}

	// Block and iterate on the main loop
	close(p.started)
	loop.Run()
	return nil
}

// handleError returns true if the error has been handled, false if the pipeline should quit
func (p *Pipeline) handleError(gErr *gst.GError) bool {
	element, reason, ok := parseDebugInfo(gErr.DebugString())
	if !ok {
		logger.Errorw("failed to parse pipeline error", errors.New(gErr.Error()),
			"debug", gErr.DebugString(),
		)
		return false
	}

	switch reason {
	case GErrNoURI, GErrCouldNotConnect:
		// bad URI or could not connect. Remove rtmp output
		if err := p.output.RemoveSinkByName(element); err != nil {
			logger.Errorw("failed to remove sink", err)
			return false
		}
		p.removed[element] = true
		return true
	case GErrFailedToStart:
		// returned after an added rtmp sink failed to start
		// should be preceded by GErrNoURI on the same sink
		return p.removed[element]
	case GErrStreamingStopped:
		// returned by queue after rtmp sink could not connect
		// should be preceded by GErrCouldNotConnect on associated sink
		if strings.HasPrefix(element, "queue_") {
			return p.removed[fmt.Sprint("sink_", element[6:])]
		}
		return false
	default:
		// input failure or file write failure. Fatal
		logger.Errorw("pipeline error", errors.New(gErr.Error()),
			"debug", gErr.DebugString(),
		)
		return false
	}
}

// Debug info comes in the following format:
// file.c(line): method_name (): /GstPipeline:pipeline/GstBin:bin_name/GstElement:element_name:\nError message
func parseDebugInfo(debug string) (element string, reason string, ok bool) {
	end := strings.Index(debug, ":\n")
	if end == -1 {
		return
	}
	start := strings.LastIndex(debug[:end], ":")
	if start == -1 {
		return
	}
	element = debug[start+1 : end]
	reason = debug[end+2:]
	if strings.HasPrefix(reason, GErrCouldNotConnect) {
		reason = GErrCouldNotConnect
	}
	ok = true
	return
}

func (p *Pipeline) AddOutput(url string) error {
	return p.output.AddRtmpSink(url)
}

func (p *Pipeline) RemoveOutput(url string) error {
	return p.output.RemoveRtmpSink(url)
}

func (p *Pipeline) Close() {
	<-p.started
	select {
	case <-p.closed:
		return
	default:
		close(p.closed)
		logger.Debugw("sending EOS to pipeline")
		p.pipeline.SendEvent(gst.NewEOSEvent())
	}
}

func requireLink(src, sink *gst.Pad) error {
	if linkReturn := src.Link(sink); linkReturn != gst.PadLinkOK {
		return fmt.Errorf("pad link: %s", linkReturn.String())
	}
	return nil
}
