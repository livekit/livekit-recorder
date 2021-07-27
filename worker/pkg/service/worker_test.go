package service

import (
	"context"
	"testing"
	"time"

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

	svc := recorder.NewRecordingService(rc)

	t.Run("Submit", func(t *testing.T) {
		id := "test-submit"
		_ = rc.Del(ctx, worker.getKey(id)).Err()
		submit(t, svc, worker, id)
		// wait to finish
		time.Sleep(time.Millisecond * 5100)
		require.Equal(t, Available, worker.status)
	})

	t.Run("Reserved", func(t *testing.T) {
		id1 := "test-reserved-1"
		id2 := "test-reserved-2"
		_ = rc.Del(ctx, worker.getKey(id1)).Err()
		_ = rc.Del(ctx, worker.getKey(id2)).Err()
		submit(t, svc, worker, id1)
		submitReserved(t, svc, id2)
		// wait to finish
		time.Sleep(time.Millisecond * 5100)
		require.Equal(t, Available, worker.status)
	})

	t.Run("Stop", func(t *testing.T) {
		id := "test-stop"
		_ = rc.Del(ctx, worker.getKey(id)).Err()
		submit(t, svc, worker, id)
		// server ends recording
		require.NoError(t, svc.EndRecording(id))
		time.Sleep(time.Millisecond * 50)
		// check that recording has ended early
		require.Equal(t, Available, worker.status)
	})

	t.Run("Kill", func(t *testing.T) {
		id := "test-kill"
		_ = rc.Del(ctx, worker.getKey(id)).Err()
		submit(t, svc, worker, id)
		// worker is killed
		worker.Stop()
		time.Sleep(time.Millisecond * 50)
		// check that recording has ended early
		require.Equal(t, Available, worker.status)
	})
}

func submit(t *testing.T, svc *recorder.RecordingService, worker *Worker, id string) {
	// send recording reservation
	msg := createRequest(t, id)

	// server sends reservation
	require.NoError(t, svc.ReserveRecording(msg, id))

	// check that worker is reserved
	require.Equal(t, Reserved, worker.status)

	// start recording
	require.NoError(t, svc.StartRecording(id))
	time.Sleep(time.Millisecond * 50)

	// check that worker is recording
	require.Equal(t, Recording, worker.status)
}

func submitReserved(t *testing.T, svc *recorder.RecordingService, id string) {
	// send recording reservation
	msg := createRequest(t, id)

	// server sends reservation
	require.Error(t, svc.ReserveRecording(msg, id))
}

func createRequest(t *testing.T, id string) string {
	req := &livekit.RecordingReservation{
		Id:          id,
		SubmittedAt: time.Now().UnixNano(),
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
	return string(b)
}
