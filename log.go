package main

import (
	"os"

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
			TimestampFormat: "15:04:05.999",
		},
	}
}
