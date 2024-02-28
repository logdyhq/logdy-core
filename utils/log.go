package utils

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

var Logger *log.Logger

func init() {
	InitLogger()
}

func InitLogger() {
	Logger = &log.Logger{
		Out:   ioutil.Discard,
		Level: log.DebugLevel,
		Formatter: &log.TextFormatter{
			DisableColors:   false,
			FullTimestamp:   true,
			TimestampFormat: "15:04:05.999",
		},
	}
}
