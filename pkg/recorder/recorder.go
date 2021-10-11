package recorder

import (
	"errors"
	"strings"
	"time"

	"github.com/livekit/protocol/logger"
	livekit "github.com/livekit/protocol/proto"

	"github.com/livekit/livekit-recorder/pkg/config"
	"github.com/livekit/livekit-recorder/pkg/display"
	"github.com/livekit/livekit-recorder/pkg/pipeline"
)

type Recorder struct {
	conf *config.Config

	req      *livekit.StartRecordingRequest
	inputUrl string

	display  *display.Display
	pipeline *pipeline.Pipeline
}

func NewRecorder(conf *config.Config) *Recorder {
	return &Recorder{
		conf: conf,
	}
}

func (r *Recorder) Validate(req *livekit.StartRecordingRequest) error {
	r.conf.ApplyDefaults(req)

	// validate input
	url, err := r.getInputUrl(req)
	if err != nil {
		return err
	}

	// validate output
	switch req.Output.(type) {
	case *livekit.StartRecordingRequest_S3Url:
		s3 := req.Output.(*livekit.StartRecordingRequest_S3Url).S3Url
		idx := strings.LastIndex(s3, "/")
		if idx < 6 ||
			!strings.HasPrefix(s3, "s3://") ||
			!strings.HasSuffix(s3, ".mp4") {
			return errors.New("s3 output must be s3://bucket/{path/}filename.mp4")
		}
	case *livekit.StartRecordingRequest_Rtmp:
	case *livekit.StartRecordingRequest_File:
		filename := req.Output.(*livekit.StartRecordingRequest_File).File
		if !strings.HasSuffix(filename, ".mp4") {
			return errors.New("file output must be {path/}filename.mp4")
		}
	default:
		return errors.New("missing output")
	}

	r.req = req
	r.inputUrl = url
	return nil
}

// Run blocks until completion
func (r *Recorder) Run(recordingId string) *livekit.RecordingResult {
	r.display = display.New()
	options := r.req.Options
	err := r.display.Launch(r.inputUrl, int(options.Width), int(options.Height), int(options.Depth))

	res := &livekit.RecordingResult{Id: recordingId}
	if r.req == nil {
		res.Error = "recorder not initialized"
		return res
	}

	r.pipeline, err = r.getPipeline(r.req)
	if err != nil {
		logger.Errorw("error building pipeline", err)
		res.Error = err.Error()
		return res
	}

	start := time.Now()
	err = r.pipeline.Start()
	if err != nil {
		logger.Errorw("error running pipeline", err)
		res.Error = err.Error()
		return res
	}
	res.Duration = time.Since(start).Milliseconds() / 1000
	return res
}

func (r *Recorder) getPipeline(req *livekit.StartRecordingRequest) (*pipeline.Pipeline, error) {
	switch req.Output.(type) {
	case *livekit.StartRecordingRequest_Rtmp:
		return pipeline.NewRtmpPipeline(req.Output.(*livekit.StartRecordingRequest_Rtmp).Rtmp.Urls, req.Options)
	case *livekit.StartRecordingRequest_S3Url:
		s3Url := req.Output.(*livekit.StartRecordingRequest_S3Url).S3Url
		return pipeline.NewS3Pipeline(r.conf.S3.Region, s3Url, req.Options)
	case *livekit.StartRecordingRequest_File:
		return pipeline.NewFilePipeline(req.Output.(*livekit.StartRecordingRequest_File).File, req.Options)
	}
	return nil, errors.New("unsupported output")
}

func (r *Recorder) AddOutput(url string) error {
	logger.Debugw("Add Output", "url", url)
	if r.pipeline == nil {
		return errors.New("missing pipeline")
	}
	return r.pipeline.AddOutput(url)
}

func (r *Recorder) RemoveOutput(url string) error {
	logger.Debugw("Remove Output", "url", url)
	if r.pipeline == nil {
		return errors.New("missing pipeline")
	}
	return r.pipeline.RemoveOutput(url)
}

func (r *Recorder) Stop() {
	if p := r.pipeline; p != nil {
		p.Close()
	}
}

// should only be called after pipeline completes
func (r *Recorder) Close() {
	r.display.Close()
}
