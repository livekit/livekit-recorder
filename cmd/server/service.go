package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/livekit/protocol/logger"
	"github.com/urfave/cli/v2"

	"github.com/livekit/livekit-recorder/pkg/messaging"
	"github.com/livekit/livekit-recorder/pkg/service"
)

func runService(c *cli.Context) error {
	conf, err := getConfig(c)
	if err != nil {
		return err
	}

	initLogger(conf.LogLevel)

	rc, err := messaging.NewMessageBus(conf)
	if err != nil {
		return err
	}
	svc, err := service.NewService(conf, rc)
	if err != nil {
		return err
	}

	if conf.HealthPort != 0 {
		go http.ListenAndServe(fmt.Sprintf(":%d", conf.HealthPort), &handler{svc: svc})
	}

	finishChan := make(chan os.Signal, 1)
	signal.Notify(finishChan, syscall.SIGTERM, syscall.SIGQUIT)

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT)

	go func() {
		select {
		case sig := <-finishChan:
			logger.Infow("Exit requested, finishing recording then shutting down", "signal", sig)
			svc.Stop(false)
		case sig := <-stopChan:
			logger.Infow("Exit requested, stopping recording and shutting down", "signal", sig)
			svc.Stop(true)
		}
	}()

	return svc.Start()
}

type handler struct {
	svc *service.Service
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(h.svc.Status()))
}
