package recorder

import (
	"errors"
	"strings"
	"time"

	"github.com/livekit/protocol/logger"
	livekit "github.com/livekit/protocol/proto"
	"github.com/tinyzimmer/go-gst/gst"

	"github.com/livekit/livekit-recorder/pkg/config"
	"github.com/livekit/livekit-recorder/pkg/display"
	"github.com/livekit/livekit-recorder/pkg/pipeline"
)

type Recorder struct {
	conf *config.Config

	isStream bool
	filename string
	display  *display.Display
	pipeline *pipeline.Pipeline
}

func NewRecorder(conf *config.Config) *Recorder {
	return &Recorder{conf: conf}
}

func (r *Recorder) Init(req *livekit.StartRecordingRequest) error {
	config.UpdateRequestParams(r.conf, req)

	// validate input
	_, err := r.getInputUrl(req)
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
		r.filename = s3[idx+1:]
		r.isStream = false
	case *livekit.StartRecordingRequest_Rtmp:
		r.isStream = true
	case *livekit.StartRecordingRequest_File:
		filename := req.Output.(*livekit.StartRecordingRequest_File).File
		if !strings.HasSuffix(filename, ".mp4") {
			return errors.New("file output must be {path/}filename.mp4")
		}
		r.filename = filename
		r.isStream = false
	default:
		return errors.New("missing output")
	}

	if r.conf.Test {
		return nil
	}

	r.display = display.New()
	return r.display.Launch("https://www.youtube.com/watch?v=v4I-19YAi2A", 1920, 1080, 24)
}

func (r *Recorder) Run(recordingId string, req *livekit.StartRecordingRequest) *livekit.RecordingResult {
	var err error
	res := &livekit.RecordingResult{Id: recordingId}

	gst.Init(nil)
	r.pipeline, err = r.getPipeline(req)
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

	if s3, ok := req.Output.(*livekit.StartRecordingRequest_S3Url); ok && !r.conf.Test {
		if err = r.upload(s3.S3Url); err != nil {
			res.Error = err.Error()
			return res
		}
		res.DownloadUrl = s3.S3Url
	}

	return res
}

func (r *Recorder) getPipeline(req *livekit.StartRecordingRequest) (*pipeline.Pipeline, error) {
	switch req.Output.(type) {
	case *livekit.StartRecordingRequest_Rtmp:
		return pipeline.NewRtmpPipeline(req.Output.(*livekit.StartRecordingRequest_Rtmp).Rtmp.Urls, req.Options)
	case *livekit.StartRecordingRequest_S3Url, *livekit.StartRecordingRequest_File:
		return pipeline.NewFilePipeline(r.filename, req.Options)
	}
	return nil, errors.New("output missing")
}

func (r *Recorder) AddOutput(url string) error {
	if !r.isStream {
		return errors.New("cannot add stream output to file recording")
	}
	return nil
}

func (r *Recorder) RemoveOutput(url string) error {
	if !r.isStream {
		return errors.New("cannot add stream output to file recording")
	}
	return nil
}

func (r *Recorder) Stop() {
	if p := r.pipeline; p != nil {
		p.Close()
	}
}

// should only be called after pipeline completes
func (r *Recorder) Close() {
	r.display.Close()
	r.display = nil
}
