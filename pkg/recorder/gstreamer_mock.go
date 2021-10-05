// +build test

package recorder

import (
	"time"

	livekit "github.com/livekit/protocol/proto"
)

type Pipeline struct {
	kill chan struct{}
}

func (p *Pipeline) Close() {
	p.kill <- struct{}{}
}

func (r *Recorder) runGStreamer(req *livekit.StartRecordingRequest) error {
	r.pipeline = &Pipeline{
		kill: make(chan struct{}, 1),
	}
	select {
	case <-time.After(time.Second * 3):
	case <-r.pipeline.kill:
	}
	return nil
}
