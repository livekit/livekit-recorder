// +build test

package recorder

import (
	livekit "github.com/livekit/protocol/proto"
)

type Pipeline struct{}

func (p *Pipeline) Close() {}

func (r *Recorder) runGStreamer(req *livekit.StartRecordingRequest) error {
	return nil
}
