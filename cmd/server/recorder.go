package main

import (
	"github.com/urfave/cli/v2"

	"github.com/livekit/livekit-recorder/pkg/recorder"
)

func runRecorder(c *cli.Context) error {
	conf, err := getConfig(c)
	if err != nil {
		return err
	}
	req, err := getRequest(c)
	if err != nil {
		return err
	}

	initLogger(conf.LogLevel)

	rec, err := recorder.NewRecorder(conf)
	if err != nil {
		return err
	}

	return rec.Start(req)
}
