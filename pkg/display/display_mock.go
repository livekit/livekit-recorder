// +build test

package display

import (
	"os/exec"
)

func LaunchXvfb(width, height, depth int) (*exec.Cmd, error) {
	return nil, nil
}

func LaunchChrome(url string, width, height int) (func(), error) {
	return nil, nil
}
