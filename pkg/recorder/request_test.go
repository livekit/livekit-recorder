package recorder

import (
	"strings"
	"testing"

	livekit "github.com/livekit/protocol/proto"
	"github.com/stretchr/testify/require"

	"github.com/livekit/livekit-recorder/pkg/config"
)

func TestInputUrl(t *testing.T) {
	req := &livekit.StartRecordingRequest{
		Input: &livekit.StartRecordingRequest_Template{
			Template: &livekit.RecordingTemplate{
				Layout:   "speaker-light",
				RoomName: "hello",
			},
		},
	}

	expected := "https://recorder.livekit.io/#/speaker-light?url=wss%3A%2F%2Ftest.livekit.cloud&token="
	rec := &Recorder{
		conf: &config.Config{
			ApiKey:          "apikey",
			ApiSecret:       "apisecret",
			WsUrl:           "wss://test.livekit.cloud",
			TemplateAddress: "https://recorder.livekit.io",
		},
	}

	actual, err := rec.GetInputUrl(req)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(actual, expected))
}
