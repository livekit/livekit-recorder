package main

import (
	"os"

	"github.com/urfave/cli/v2"

	"github.com/livekit/livekit-recorder/pkg/recorder"
)

func runRecorder(c *cli.Context) error {
	initLogger("debug")

	if err := os.Setenv("DISPLAY", recorder.Display); err != nil {
		return err
	}

	rec := recorder.NewRecorder(nil)
	rec.Start()
	return nil
}
