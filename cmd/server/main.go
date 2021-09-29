package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-logr/zapr"
	"github.com/livekit/protocol/logger"
	livekit "github.com/livekit/protocol/proto"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/livekit/livekit-recorder/pkg/config"
	"github.com/livekit/livekit-recorder/version"
)

func main() {
	app := &cli.App{
		Name:        "livekit-recorder-service",
		Usage:       "LiveKit Recorder Service",
		Description: "runs the recording service",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Usage: "path to LiveKit recording config defaults",
			},
			&cli.StringFlag{
				Name:    "config-body",
				Usage:   "Default LiveKit recording config in JSON, typically passed in as an env var in a container",
				EnvVars: []string{"LIVEKIT_RECORDER_SVC_CONFIG"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "start-recording",
				Usage:  "Starts a controller server",
				Action: runRecorder,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "request",
						Usage: "path to json StartRecordingRequest file",
					},
					&cli.StringFlag{
						Name:  "request-body",
						Usage: "StartRecordingRequest json",
					},
				},
			},
			{
				Name:   "start-service",
				Usage:  "Starts an origin server",
				Action: runService,
			},
		},
		Version: version.Version,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

func getConfig(c *cli.Context) (*config.Config, error) {
	configFile := c.String("config")
	configBody := c.String("config-body")
	if configBody == "" {
		if configFile != "" {
			content, err := ioutil.ReadFile(configFile)
			if err != nil {
				return nil, err
			}
			configBody = string(content)
		} else {
			return nil, errors.New("missing config")
		}
	}

	return config.NewConfig(configBody)
}

func getRequest(c *cli.Context) (*livekit.StartRecordingRequest, error) {
	reqFile := c.String("request")
	reqBody := c.String("request-body")

	var content []byte
	var err error
	if reqBody != "" {
		content = []byte(reqBody)
	} else if reqFile != "" {
		content, err = ioutil.ReadFile(reqFile)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("missing request")
	}

	req := &livekit.StartRecordingRequest{}
	err = protojson.Unmarshal(content, req)
	return req, err
}

func initLogger(level string) {
	conf := zap.NewProductionConfig()
	if level != "" {
		lvl := zapcore.Level(0)
		if err := lvl.UnmarshalText([]byte(level)); err == nil {
			conf.Level = zap.NewAtomicLevelAt(lvl)
		}
	}

	l, _ := conf.Build()
	logger.SetLogger(zapr.NewLogger(l), "livekit-recorder")
}
