//go:build test
// +build test

package display

import (
	"github.com/livekit/protocol/livekit"

	"github.com/livekit/livekit-recorder/pkg/config"
)

type Display struct {
	startChan chan struct{}
	endChan   chan struct{}
}

func Launch(conf *config.Config, url string, opts *livekit.RecordingOptions, isTemplate bool) (*Display, error) {
	return &Display{
		startChan: make(chan struct{}),
		endChan:   make(chan struct{}),
	}, nil
}

func (d *Display) RoomStarted() chan struct{} {
	return d.startChan
}

func (d *Display) RoomEnded() chan struct{} {
	return d.endChan
}

func (d *Display) Close() {
	close(d.endChan)
}
