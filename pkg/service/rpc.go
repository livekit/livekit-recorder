package service

import (
	"fmt"

	"github.com/livekit/protocol/logger"
	livekit "github.com/livekit/protocol/proto"
	"github.com/livekit/protocol/recording"
	"google.golang.org/protobuf/proto"
)

func (s *Service) handleRecording() {
	// subscribe to request channel
	requests, err := s.bus.Subscribe(s.ctx, recording.RequestChannel(s.recordingId))
	if err != nil {
		return
	}
	defer requests.Close()

	// ready to accept requests
	err = s.handleResponse(s.recordingId, s.recordingId, nil)
	if err != nil {
		return
	}

	// listen for rpcs
	logger.Debugw("Waiting for requests", "recordingId", s.recordingId)
	result := make(chan *livekit.RecordingResult, 1)
	for {
		select {
		case <-s.kill:
			// kill signal received, stop recorder
			if status := s.status.Load(); status != Stopping {
				s.status.Store(Stopping)
				s.rec.Stop()
			}
		case res := <-result:
			// recording stopped, send results to result channel
			b, err := proto.Marshal(res)
			if err != nil {
				logger.Errorw("Failed to marshal results", err)
			} else if err = s.bus.Publish(s.ctx, recording.ResultChannel, b); err != nil {
				logger.Errorw("Failed to write results", err)
			}

			// clean up
			if err = s.rec.Close(); err != nil {
				logger.Errorw("Failed to close recorder", err)
			}
			return
		case msg := <-requests.Channel():
			// unmarshal request
			req := &livekit.RecordingRequest{}
			err = proto.Unmarshal(requests.Payload(msg), req)
			if err != nil {
				logger.Errorw("Failed to read request", err, "recordingId", s.recordingId)
				continue
			}

			s.handleRequest(req, result)
		}
	}
}

func (s *Service) handleRequest(req *livekit.RecordingRequest, result chan *livekit.RecordingResult) {
	var err error
	switch req.Request.(type) {
	case *livekit.RecordingRequest_Start:
		if status := s.status.Load(); status != Reserved {
			err = fmt.Errorf("tried calling start with state %s", status)
			break
		}

		// launch recorder
		s.status.Store(Starting)
		start := req.Request.(*livekit.RecordingRequest_Start).Start
		err = s.rec.Init(start)
		if err != nil {
			// failed to start, close recorder
			result <- &livekit.RecordingResult{
				Id:    s.recordingId,
				Error: err.Error(),
			}
			break
		}

		go func() {
			// blocks until recorder is finished
			result <- s.rec.Run(s.recordingId, start)
		}()
		s.status.Store(Recording)
	case *livekit.RecordingRequest_AddOutput:
		if status := s.status.Load(); status != Recording {
			err = fmt.Errorf("tried calling AddOutput with status %s", status)
			break
		}
		err = s.rec.AddOutput(req.Request.(*livekit.RecordingRequest_AddOutput).AddOutput.RtmpUrl)
	case *livekit.RecordingRequest_RemoveOutput:
		if status := s.status.Load(); status != Recording {
			err = fmt.Errorf("tried calling RemoveOutput with status %s", status)
			break
		}
		err = s.rec.RemoveOutput(req.Request.(*livekit.RecordingRequest_RemoveOutput).RemoveOutput.RtmpUrl)
	case *livekit.RecordingRequest_End:
		if status := s.status.Load(); status != Recording {
			err = fmt.Errorf("tried calling End with status %s", status)
			break
		}
		s.status.Store(Stopping)
		s.rec.Stop()
	}

	_ = s.handleResponse(s.recordingId, req.RequestId, err)
}

func (s *Service) handleResponse(recordingId, requestId string, err error) error {
	var message string
	if err != nil {
		logger.Errorw("Error handling request", err,
			"recordingId", recordingId, "requestId", requestId)
		message = err.Error()
	}

	b, err := proto.Marshal(&livekit.RecordingResponse{
		RequestId: requestId,
		Success:   err == nil,
		Error:     message,
	})
	if err != nil {
		return err
	}

	return s.bus.Publish(s.ctx, recording.ResponseChannel(recordingId), b)
}
