package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"github.com/livekit/livekit-recording/worker/pkg/config"
	"github.com/livekit/livekit-recording/worker/pkg/logger"
	"github.com/livekit/livekit-recording/worker/pkg/service"
	"github.com/livekit/livekit-recording/worker/version"
)

func main() {
	app := &cli.App{
		Name:        "livekit-recording-worker",
		Usage:       "LiveKit recording worker",
		Description: "runs the recording worker",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Usage: "path to LiveKit config file",
			},
			&cli.StringFlag{
				Name:    "config-body",
				Usage:   "LiveKit config in YAML, typically passed in as an environment var in a container",
				EnvVars: []string{"LIVEKIT_RECORDING_CONFIG"},
			},
			&cli.StringFlag{
				Name:    "redis-host",
				Usage:   "host (incl. port) to redis server",
				EnvVars: []string{"REDIS_HOST"},
			},
			&cli.StringFlag{
				Name:    "redis-password",
				Usage:   "password to redis",
				EnvVars: []string{"REDIS_PASSWORD"},
			},
		},
		Action:  startWorker,
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

func startWorker(c *cli.Context) error {
	conf, err := getConfig(c)
	if err != nil {
		return err
	}

	logger.Init(conf.LogLevel)

	// redis work queue
	logger.Infow("connecting to redis work queue", "addr", conf.Redis.Address)
	rc := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Address,
		Username: conf.Redis.Username,
		Password: conf.Redis.Password,
		DB:       conf.Redis.DB,
	})
	if err := rc.Ping(context.Background()).Err(); err != nil {
		err = errors.Wrap(err, "unable to connect to redis")
		return err
	}

	worker := service.InitializeWorker(rc)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		sig := <-sigChan
		logger.Infow("exit requested, shutting down", "signal", sig)
		worker.Stop()
	}()

	return worker.Start()
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
