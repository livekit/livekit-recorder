package service

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/livekit/protocol/logger"
	livekit "github.com/livekit/protocol/proto"
	"github.com/livekit/protocol/utils"
	"google.golang.org/protobuf/proto"

	"github.com/livekit/livekit-recorder/pkg/config"
	"github.com/livekit/livekit-recorder/pkg/recorder"
)

type Service struct {
	rec      *recorder.Recorder
	ctx      context.Context
	bus      utils.MessageBus
	status   atomic.Value // Status
	shutdown chan struct{}
	kill     chan struct{}

	mock bool
}

type Status string

const (
	Available Status = "available"
	Reserved  Status = "reserved"
	Recording Status = "recording"
)

func NewService(conf *config.Config, bus utils.MessageBus) (*Service, error) {
	rec, err := recorder.NewRecorder(conf)
	if err != nil {
		return nil, err
	}

	return &Service{
		rec:      rec,
		ctx:      context.Background(),
		bus:      bus,
		status:   atomic.Value{},
		shutdown: make(chan struct{}, 1),
		kill:     make(chan struct{}, 1),
		mock:     conf.Test,
	}, nil
}

func (w *Service) Start() error {
	logger.Debugw("Starting worker", "mock", w.mock)

	reservations, err := w.bus.SubscribeQueue(context.Background(), utils.ReservationChannel)
	if err != nil {
		return err
	}
	defer reservations.Close()

	for {
		w.status.Store(Available)
		logger.Debugw("Recorder waiting")

		select {
		case <-w.shutdown:
			logger.Debugw("Shutting down")
			return nil
		case msg := <-reservations.Channel():
			logger.Debugw("Request received")

			req := &livekit.RecordingReservation{}
			err := proto.Unmarshal(reservations.Payload(msg), req)
			if err != nil {
				logger.Errorw("Malformed request", err)
				continue
			}

			if req.SubmittedAt < time.Now().Add(-utils.ReservationTimeout).UnixNano() {
				logger.Debugw("Discarding old request", "ID", req.Id)
				continue
			}

			w.status.Store(Reserved)
			logger.Debugw("Request claimed", "ID", req.Id)

			res, err := w.Record(req)
			b, _ := proto.Marshal(res)
			_ = w.bus.Publish(w.ctx, utils.RecordingResultChannel, b)
			if err != nil {
				return err
			}
		}
	}
}

func (w *Service) Record(req *livekit.RecordingReservation) (res *livekit.RecordingResult, err error) {
	res = &livekit.RecordingResult{Id: req.Id}
	var startedAt time.Time
	defer func() {
		if err != nil {
			logger.Errorw("Recorder failed", err)
			res.Error = err.Error()
		} else {
			res.Duration = time.Since(startedAt).Milliseconds()
		}
	}()

	start, err := w.bus.Subscribe(w.ctx, utils.StartRecordingChannel(req.Id))
	if err != nil {
		return
	}
	defer start.Close()

	stop, err := w.bus.Subscribe(w.ctx, utils.EndRecordingChannel(req.Id))
	if err != nil {
		return
	}
	defer stop.Close()

	err = w.bus.Publish(w.ctx, utils.ReservationResponseChannel(req.Id), nil)
	if err != nil {
		return
	}

	// send recording started message
	<-start.Channel()
	w.status.Store(Recording)

	// conf, err := config.Merge(w.defaults, req)
	// if err != nil {
	// 	return
	// }

	// Launch node recorder
	// done, err := recorder.StartRecording(req.Request)
	// if err != nil {
	// 	return
	// }
	// startedAt = time.Now()
	// logger.Infow("Recording started", "ID", req.Id)

	// select {
	// case err = <-done:
	// 	break
	// case <-stop.Channel():
	// 	logger.Infow("Recording stopped by livekit server", "ID", req.Id)
	// 	err = cmd.Process.Signal(syscall.SIGTERM)
	// case <-w.kill:
	// 	logger.Infow("Recording stopped by recording service interrupt", "ID", req.Id)
	// 	err = cmd.Process.Signal(syscall.SIGTERM)
	// }

	return
}

func (w *Service) Status() Status {
	return w.status.Load().(Status)
}

func (w *Service) Stop(kill bool) {
	w.shutdown <- struct{}{}
	if kill {
		w.kill <- struct{}{}
	}
}

func (w *Service) getKey(id string) string {
	return fmt.Sprintf("recording-lock-%s", id)
}
