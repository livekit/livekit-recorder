package service

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/livekit/livekit-server/pkg/recorder"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/livekit/livekit-recording/worker/pkg/config"
	"github.com/livekit/livekit-recording/worker/pkg/logger"
	livekit "github.com/livekit/livekit-recording/worker/proto"
)

func TestWorker(t *testing.T) {
	logger.Init("debug")

	ctx := context.Background()
	conf := config.TestConfig()
	rc, err := StartRedis(conf)
	require.NoError(t, err)

	worker := InitializeWorker(conf, rc)
	go func() {
		err := worker.Start()
		require.NoError(t, err)
	}()

	testSubmit(t, ctx, worker, rc, "test-recording")
	// wait to finish
	time.Sleep(time.Second * 2)
	require.Equal(t, Available, worker.status)
}

func testSubmit(t *testing.T, ctx context.Context, worker *Worker, rc *redis.Client, id string) {
	// send recording reservation
	req := &livekit.StartRoomRecording{
		Id: id,
		Input: &livekit.RecordingInput{
			Template: &livekit.RecordingTemplate{
				Type:  "grid",
				WsUrl: "wss://testing.livekit.io",
				Token: "token",
			},
			Framerate: 60,
		},
		Output: &livekit.RecordingOutput{
			File:         "recording.mp4",
			VideoBitrate: "1000k",
			VideoBuffer:  "2000k",
		},
	}
	b, err := proto.Marshal(req)
	require.NoError(t, err)

	resp := rc.Subscribe(ctx, recorder.ResponseChannel(req.Id)).Channel()
	require.NoError(t, rc.Publish(ctx, recorder.ReservationChannel, b).Err())

	// get response from worker
	select {
	case <-resp:
	case <-time.After(time.Second):
		t.Error("no response from worker")
	}
	// check that worker is reserved
	require.Equal(t, Reserved, worker.status)

	// start recording
	require.NoError(t, rc.Publish(ctx, recorder.StartRecordingChannel(req.Id), b).Err())
	time.Sleep(time.Millisecond * 50)
	require.Equal(t, Recording, worker.status)
}
