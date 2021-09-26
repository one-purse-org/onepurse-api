package main

import (
	"github.com/isongjosiah/work/onepurse-api/config"
	"github.com/sirupsen/logrus"
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


}
