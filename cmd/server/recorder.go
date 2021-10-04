package main

import (
	"errors"

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

	rec := recorder.NewRecorder(conf)
	res := rec.Run("", req)
	if res.Error == "" {
		return nil
	}
	return errors.New(res.Error)
}
