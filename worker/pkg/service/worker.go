package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/livekit/livekit-server/pkg/recorder"
	"github.com/pkg/errors"
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
	status   Status
	mock     bool
}

type Status string

const (
	Available Status = "available"
	Reserved  Status = "reserved"
	Recording Status = "recording"
)

func InitializeWorker(conf *config.Config, rc *redis.Client) *Worker {
	return &Worker{
		ctx:      context.Background(),
		rc:       rc,
		defaults: conf,
		kill:     make(chan struct{}),
		status:   Available,
		mock:     conf.Test,
	}
}

func (w *Worker) Start() error {
	logger.Debugw("Starting worker")

	pubsub := w.rc.Subscribe(w.ctx, recorder.ReservationChannel)
	for msg := range pubsub.Channel() {
		logger.Debugw("Request received")
		req := &livekit.StartRoomRecording{}
		err := proto.Unmarshal([]byte(msg.Payload), req)
		if err != nil {
			return err
		}

		key := w.getKey(req)
		claimed, start, stop, err := w.Claim(req.Id, key)
		if err != nil {
			logger.Errorw("Request failed", err)
			return err
		} else if !claimed {
			logger.Debugw("Request locked")
			continue
		}
		logger.Debugw("Request claimed")

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

	w.status = Reserved
	start = w.rc.Subscribe(w.ctx, recorder.StartRecordingChannel(id)).Channel()
	kill = w.rc.Subscribe(w.ctx, recorder.EndRecordingChannel(id)).Channel()
	err = w.rc.Publish(w.ctx, recorder.ResponseChannel(id), nil).Err()
	if err != nil {
		_ = w.rc.Del(w.ctx, key).Err()
	}
	return
}

func (w *Worker) Run(req *livekit.StartRoomRecording, key string, start, stop <-chan *redis.Message) error {
	<-start
	w.status = Recording

	defer func() {
		err := w.rc.Del(w.ctx, key).Err()
		if err != nil {
			logger.Errorw("failed to unlock job", err)
		}
		w.status = Available
	}()

	// Launch node recorder
	done := make(chan error)
	go w.Launch(req, done)

	select {
	case err := <-done:
		if err != nil {
			logger.Errorw("Recording failed", err)
		} else {
			logger.Infow("Recording finished")
		}
	case <-stop:
		logger.Infow("Recording stopped by livekit server")
		// TODO: kill
	case <-w.kill:
		logger.Infow("Recording killed by recording service")
		// TODO: kill
	}

	return nil
}

func (w *Worker) Launch(req *livekit.StartRoomRecording, done chan error) {
	_, err := config.Merge(w.defaults, req)
	if err != nil {
		done <- errors.Wrap(err, "failed to build recorder config")
		return
	}

	logger.Debugw("Recording started")
	if w.mock {
		time.Sleep(time.Second * 1)
	} else {
		// TODO: launch
	}

	done <- nil
}

func (w *Worker) Stop() {
	w.kill <- struct{}{}
}

func (w *Worker) getKey(recording *livekit.StartRoomRecording) string {
	return fmt.Sprintf("recording-%s", recording.Id)
}
