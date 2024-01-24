package main

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

var logger *log.Logger

func initLogger() {
	logger = &log.Logger{
		Out:   os.Stdout,
		Level: log.DebugLevel,
		Formatter: &log.TextFormatter{
			DisableColors:   false,
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339Nano,
		},
	}
}
