// +build integration

package test

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"testing"

	livekit "github.com/livekit/protocol/proto"
	"github.com/stretchr/testify/require"

	"github.com/livekit/livekit-recorder/pkg/config"
	"github.com/livekit/livekit-recorder/pkg/recorder"
)

func TestRecorder(t *testing.T) {
	conf, err := config.NewConfig("")
	require.NoError(t, err)
	conf.ApiKey = "key"
	conf.ApiSecret = "secret"
	conf.WsUrl = "ws://localhost:7880"

	if !t.Run("template-test", func(t *testing.T) {
		runTemplateTest(t, conf)
	}) {
		t.FailNow()
	}

	// if !t.Run("rtmp-test", func(t *testing.T) {
	// 	runRtmpTest(t, conf)
	// }) {
	// 	t.FailNow()
	// }
}

func runTemplateTest(t *testing.T, conf *config.Config) {
	filename := "/Users/dc/Downloads/s3-test.mp4"
	req := &livekit.StartRecordingRequest{
		Input: &livekit.StartRecordingRequest_Template{
			Template: &livekit.RecordingTemplate{
				Layout: "speaker-dark",
				Room: &livekit.RecordingTemplate_RoomName{
					RoomName: "recorder-test",
				},
			},
		},
		Output: &livekit.StartRecordingRequest_Filepath{
			Filepath: filename,
		},
	}

	// TODO: start publishing video to room

	rec := recorder.NewRecorder(conf)
	require.NoError(t, rec.Validate(req))

	// // record for 15s
	// time.AfterFunc(time.Second*15, func() {
	// 	rec.Stop()
	// })
	// res := rec.Run("room_test")
	//
	// // check error
	// require.Empty(t, res.Error)

	info, err := ffprobe(filename)
	require.NoError(t, err, "ffprobe failed")

	require.NotEqual(t, 0, info.Format.Size)
	// TODO: compare duration to res.Duration
	require.NotEqual(t, 0, info.Format.Duration)
	require.Equal(t, "x264", info.Format.Tags.Encoder)
	require.Equal(t, 100, info.Format.ProbeScore)
	require.Len(t, info.Streams, 2)

	var hasAudio, hasVideo bool
	for _, stream := range info.Streams {
		switch stream.CodecType {
		case "audio":
			hasAudio = true
			require.Equal(t, "aac", stream.CodecName)
			require.Equal(t, 2, stream.Channels)
			require.Equal(t, "stereo", stream.ChannelLayout)
			require.Equal(t, fmt.Sprint(req.Options.AudioFrequency), stream.SampleRate)
			// TODO: compare bitrate to req.Options.AudioBitrate "bit_rate": "135495",
			require.NotEqual(t, 0, stream.BitRate)
		case "video":
			hasVideo = true
			require.Equal(t, "h264", stream.CodecName)
			require.Equal(t, req.Options.Width, stream.Width)
			require.Equal(t, req.Options.Height, stream.Height)
			require.Equal(t, fmt.Sprintf("%d/1", req.Options.Framerate), stream.RFrameRate)
			// TODO: compare bitrate to req.Options.VideoBitrate  "bit_rate": "3783664",
			require.NotEqual(t, 0, stream.BitRate)
		default:
			t.Fatalf("unrecognized stream type %s", stream.CodecType)
		}
	}
	require.True(t, hasAudio && hasVideo)
}

func runRtmpTest(t *testing.T, conf *config.Config) {
	req := &livekit.StartRecordingRequest{
		Input: &livekit.StartRecordingRequest_Url{
			Url: "TODO",
		},
		Output: &livekit.StartRecordingRequest_Rtmp{
			Rtmp: &livekit.RtmpOutput{
				Urls: []string{"TODO", "TODO"},
			},
		},
	}

	rec := recorder.NewRecorder(conf)
	require.NoError(t, rec.Validate(req))
	resChan := make(chan *livekit.RecordingResult, 1)
	go func() {
		resChan <- rec.Run("rtmp_test")
	}()

	// TODO: verify stream with ffprobe

	// TODO: add rtmp

	// TODO: remove rtmp

	// stop
	rec.Stop()
	res := <-resChan

	// check error
	require.Empty(t, res.Error)

	// TODO: check duration
}

type FFProbeInfo struct {
	Streams []struct {
		CodecName string `json:"codec_name"`
		CodecType string `json:"codec_type"`

		// audio
		SampleRate    string `json:"sample_rate"`
		Channels      int    `json:"channels"`
		ChannelLayout string `json:"channel_layout"`

		// video
		Width      int32  `json:"width"`
		Height     int32  `json:"height"`
		RFrameRate string `json:"r_frame_rate"`
		BitRate    string `json:"bit_rate"`
	} `json:"streams"`
	Format struct {
		Filename   string `json:"filename"`
		Duration   string `json:"duration"`
		Size       string `json:"size"`
		ProbeScore int    `json:"probe_score"`
		Tags       struct {
			Encoder string `json:"encoder"`
		} `json:"tags"`
	} `json:"format"`
}

func ffprobe(filename string) (*FFProbeInfo, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-hide_banner",
		"-show_format",
		"-show_streams",
		"-print_format", "json",
		filename,
	)

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	info := &FFProbeInfo{}
	err = json.Unmarshal(out, info)
	return info, err
}
