package utils

import (
	"io"
	"os"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

var Logger *log.Logger

type LOG_LEVEL int

type LogInterceptor func(entry *LogEntry)

type LogEntry struct {
	Data map[string]interface{}
	Time time.Time

	// Level the log entry was logged at: Trace, Debug, Info, Warn, Error, Fatal or Panic
	// This field will be set on entry firing and the value will be equal to the one in Logger struct field.
	Level int

	// Calling method, with package name
	Caller *runtime.Frame

	// Message passed to Trace, Debug, Info, Warn, Error, Fatal or Panic
	Message string
}

const LOG_LEVEL_SILENT LOG_LEVEL = 0
const LOG_LEVEL_NORMAL LOG_LEVEL = 1
const LOG_LEVEL_VERBOSE LOG_LEVEL = 2

func init() {
	InitLogger()
}

func InitLogger() {
	Logger = &log.Logger{
		Out:   io.Discard,
		Level: log.InfoLevel,
		Formatter: &log.TextFormatter{
			DisableColors:   false,
			FullTimestamp:   true,
			TimestampFormat: "15:04:05.999",
		},
		Hooks: make(log.LevelHooks),
	}
}

func SetLoggerLevel(verbose bool) {
	if verbose {
		Logger.SetLevel(logrus.TraceLevel)
		Logger.Debug("Setting verbose logger")
	} else {
		Logger.SetLevel(logrus.InfoLevel)
	}
}

func SetLoggerDiscard(discard bool) {
	if discard {
		Logger.Out = io.Discard
	} else {
		Logger.Out = os.Stdout
	}
}

func SetLoggerInterceptor(fn LogInterceptor) {
	Logger.AddHook(&CustomHook{
		fn: fn,
	})
}

type CustomHook struct {
	fn LogInterceptor
}

// Levels returns the log levels that this hook will be fired for
func (hook *CustomHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire is called whenever a log entry is made
func (hook *CustomHook) Fire(entry *logrus.Entry) error {
	hook.fn(&LogEntry{
		Data:    entry.Data,
		Time:    entry.Time,
		Level:   int(entry.Level),
		Caller:  entry.Caller,
		Message: entry.Message,
	})
	return nil
}
