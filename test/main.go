package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	livekit "github.com/livekit/protocol/proto"

	"github.com/livekit/livekit-recorder/pkg/config"
	"github.com/livekit/livekit-recorder/pkg/recorder"
)

func main() {
	err := runTests()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runTests() error {
	conf, err := config.NewConfig("")
	if err != nil {
		return err
	}

	if err = runTemplateTest(conf); err != nil {
		return err
	}

	if err = runRtmpTest(conf); err != nil {
		return err
	}

	return nil
}

func runTemplateTest(conf *config.Config) error {
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
			Filepath: "template-test.mp4",
		},
	}

	// TODO: start publishing video to room

	rec := recorder.NewRecorder(conf)
	if err := rec.Validate(req); err != nil {
		return err
	}

	// record for 15s
	time.AfterFunc(time.Second*15, func() {
		rec.Stop()
	})
	res := rec.Run("room_test")

	// check error
	if res.Error != "" {
		return errors.New(res.Error)
	}

	// TODO: check duration

	// TODO: verify file using ffprobe

	return nil
}

func runRtmpTest(conf *config.Config) error {
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
	if err := rec.Validate(req); err != nil {
		return err
	}
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
	if res.Error != "" {
		return errors.New(res.Error)
	}

	// TODO: check duration

	return nil
}
