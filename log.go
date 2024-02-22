package main

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

var logger *log.Logger

func initLogger() {
	logger = &log.Logger{
		Out:   ioutil.Discard,
		Level: log.DebugLevel,
		Formatter: &log.TextFormatter{
			DisableColors:   false,
			FullTimestamp:   true,
			TimestampFormat: "15:04:05.999",
		},
	}
}
