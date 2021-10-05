// +build !linux

package recorder

import (
	"os/exec"
)

func (r *Recorder) launchXvfb(width, height, depth int) (*exec.Cmd, error) {
	return nil, nil
}

func (r *Recorder) launchChrome(url string, width, height int) (func(), error) {
	return nil, nil
}
