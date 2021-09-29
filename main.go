package main

import (
	"github.com/isongjosiah/work/onepurse-api/api"
	"github.com/isongjosiah/work/onepurse-api/config"
	"github.com/isongjosiah/work/onepurse-api/deps"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	allowConnectionsAfterShutdown = 5 * time.Second
)

func main() {
	cfg := config.New()
	logrus.Info("[Env]: OK")

	if cfg.Environment == "development" {
		logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	} else {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	deps, err := deps.New(cfg)
	if err != nil {
		logrus.Fatalf("Unable to setup dependencies : %s", err.Error())
	}
	logrus.Info("[DEPS]: OK")

	a := &api.API{
		Config: cfg,
		Deps:   deps,
	}

	go func() {
		log.Fatal(a.Serve())
	}()

	// graceful shutdown
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-stopChan

	logrus.Infof("[API]: Request to shutdown server. Doing nothing for %v", allowConnectionsAfterShutdown)
	waitTimer := time.NewTimer(allowConnectionsAfterShutdown)
	<-waitTimer.C

	logrus.Info("[API]: Shutting down server ...")
	logrus.Fatal(a.Shutdown())
}
