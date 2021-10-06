// +build test

package pipeline

import (
	"time"

	livekit "github.com/livekit/protocol/proto"
)

type Pipeline struct {
	kill chan struct{}
}

func (p *Pipeline) Start() error {
	select {
	case <-time.After(time.Second * 3):
	case <-p.kill:
	}
	return nil
}

func (p *Pipeline) Close() {
	p.kill <- struct{}{}
}

func NewRtmpPipeline(rtmp []string, options *livekit.RecordingOptions) (*Pipeline, error) {
	return &Pipeline{
		kill: make(chan struct{}, 1),
	}, nil
}

func NewFilePipeline(filename string, options *livekit.RecordingOptions) (*Pipeline, error) {
	return &Pipeline{
		kill: make(chan struct{}, 1),
	}, nil
}
