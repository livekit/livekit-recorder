package service

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/go-redis/redis/v8"
	"github.com/livekit/livekit-server/pkg/recorder"
	"google.golang.org/protobuf/proto"

	"github.com/livekit/livekit-recording/worker/pkg/config"
	"github.com/livekit/livekit-recording/worker/pkg/logger"
	livekit "github.com/livekit/livekit-recording/worker/proto"
)

type Worker struct {
	ctx      context.Context
	rc       *redis.Client
	defaults *config.Config
	kill     chan struct{}
}

func InitializeWorker(conf *config.Config, rc *redis.Client) *Worker {
	return &Worker{
		ctx:      context.Background(),
		rc:       rc,
		defaults: conf,
		kill:     make(chan struct{}),
	}
}

func (w *Worker) Start() error {
	pubsub := w.rc.Subscribe(w.ctx, recorder.ReservationChannel)
	for msg := range pubsub.Channel() {
		req := &livekit.StartRoomRecording{}
		err := proto.Unmarshal([]byte(msg.Payload), req)
		if err != nil {
			return err
		}

		key := w.getKey(req)
		claimed, start, stop, err := w.Claim(req.Id, key)
		if err != nil {
			return err
		} else if !claimed {
			continue
		}

		err = w.Run(req, key, start, stop)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *Worker) Claim(id, key string) (locked bool, start, kill <-chan *redis.Message, err error) {
	locked, err = w.rc.SetNX(w.ctx, key, rand.Int(), 0).Result()
	if !locked || err != nil {
		return
	}

	start = w.rc.Subscribe(w.ctx, recorder.StartRecordingChannel(id)).Channel()
	kill = w.rc.Subscribe(w.ctx, recorder.EndRecordingChannel(id)).Channel()
	err = w.rc.Publish(w.ctx, recorder.ResponseChannel(id), nil).Err()
	return
}

func (w *Worker) Run(req *livekit.StartRoomRecording, key string, start, stop <-chan *redis.Message) error {
	<-start

	defer func() {
		err := w.rc.Del(w.ctx, key).Err()
		if err != nil {
			logger.Errorw("failed to unlock redis job", err)
		}
	}()

	// Launch node recorder
	done := make(chan struct{})
	go w.Launch(req, done)

	select {
	case <-done:
		logger.Infow("Recording finished")
	case <-stop:
		logger.Infow("Recording stopped by livekit server")
	case <-w.kill:
		logger.Infow("Recording killed by recording service")
	}

	return nil
}

func (w *Worker) Launch(req *livekit.StartRoomRecording, done chan struct{}) {
	// TODO: launch
	done <- struct{}{}
}

func (w *Worker) Stop() {
	w.kill <- struct{}{}
}

func (w *Worker) getKey(recording *livekit.StartRoomRecording) string {
	return fmt.Sprintf("recording-%s", recording.Id)
}
