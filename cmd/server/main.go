package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-logr/zapr"
	"github.com/livekit/protocol/logger"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

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
			&cli.StringFlag{
				Name:    "redis-host",
				Usage:   "host (incl. port) to redis server",
				EnvVars: []string{"REDIS_HOST"},
			},
		},
		Action:  runRecorder,
		Version: version.Version,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}

func getConfig(c *cli.Context) (*config.Config, error) {
	confString, err := getConfigString(c)
	if err != nil {
		return nil, err
	}

	return config.NewConfig(confString, c)
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

func getConfigString(c *cli.Context) (string, error) {
	configFile := c.String("config")
	configBody := c.String("config-body")
	if configBody == "" {
		if configFile != "" {
			content, err := ioutil.ReadFile(configFile)
			if err != nil {
				return "", err
			}
			configBody = string(content)
		}
	}
	return configBody, nil
}
