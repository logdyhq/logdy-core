package logdy

import (
	"encoding/json"
	_http "net/http"

	"github.com/logdyhq/logdy-core/http"
	"github.com/logdyhq/logdy-core/models"
	"github.com/logdyhq/logdy-core/modes"
	"github.com/logdyhq/logdy-core/utils"
)

type Config struct {
	// Whether UI events should be tracked in the UI
	AnalyticsEnabled bool

	// The passphrase to access the UI
	UiPass string

	// A path to the config to be loaded in the UI
	ConfigFilePath string

	// A time window during which log messages are gathered and send in a bulk to a client.
	// Decreasing this window will improve the 'real-time' feeling of messages presented on
	// the screen but could decrease UI performance
	BulkWindowMs int64

	// A URL path on which Logdy will be accessible
	HttpPathPrefix string

	// A server port on which the UI will be served
	ServerPort string
	// A server IP on which the UI will be served, leave empty to bind to existing server
	ServerIp        string
	MaxMessageCount int64

	// Log level
	LogLevel LOG_LEVEL

	// A function to be invoked when a Logdy internal log message is produced
	// If not nil then LogLevel is ignored
	LogInterceptor LogInterceptor

	// Key to be used when communicating with the REST API
	ApiKey string
}

type LOG_LEVEL = utils.LOG_LEVEL
type LogInterceptor = utils.LogInterceptor
type LogEntry = utils.LogEntry

const LOG_LEVEL_SILENT LOG_LEVEL = utils.LOG_LEVEL_SILENT
const LOG_LEVEL_NORMAL LOG_LEVEL = utils.LOG_LEVEL_NORMAL
const LOG_LEVEL_VERBOSE LOG_LEVEL = utils.LOG_LEVEL_VERBOSE

type Logdy interface {
	Config() *Config
	Log(fields Fields) error
	LogString(message string) error
}

type Fields map[string]interface{}

type LogdyInstance struct {
	config *Config
}

func (l *LogdyInstance) Log(fields Fields) error {

	serialized, err := json.Marshal(fields)
	if err != nil {
		return err
	}

	modes.ProduceMessageString(http.Ch, string(serialized), models.MessageTypeStdout, &models.MessageOrigin{})
	return nil
}
func (l *LogdyInstance) Config() *Config {
	return l.config
}

func (l *LogdyInstance) LogString(message string) error {
	modes.ProduceMessageString(http.Ch, message, models.MessageTypeStdout, &models.MessageOrigin{})
	return nil
}

func translateToConfig(c *Config) http.Config {
	return http.Config{
		AnalyticsEnabled: c.AnalyticsEnabled,
		UiPass:           c.UiPass,
		ConfigFilePath:   c.ConfigFilePath,
		BulkWindowMs:     c.BulkWindowMs,
		HttpPathPrefix:   c.HttpPathPrefix,
		ServerPort:       c.ServerPort,
		ServerIp:         c.ServerIp,
		MaxMessageCount:  c.MaxMessageCount,
		LogLevel:         c.LogLevel,
		LogInterceptor:   c.LogInterceptor,
		ApiKey:           c.ApiKey,
	}
}

func InitializeLogdy(config Config, serveMux *_http.ServeMux) Logdy {
	utils.InitLogger()

	switch config.LogLevel {
	case LOG_LEVEL_SILENT:
		utils.SetLoggerDiscard(true)
	case LOG_LEVEL_NORMAL:
		utils.SetLoggerDiscard(false)
		utils.SetLoggerLevel(false)
	case LOG_LEVEL_VERBOSE:
		utils.SetLoggerDiscard(false)
		utils.SetLoggerLevel(true)
	}

	if config.LogInterceptor != nil {
		utils.SetLoggerInterceptor(config.LogInterceptor)
	}

	c := translateToConfig(&config)

	http.InitChannel()
	http.InitializeClients(c)
	http.HandleHttp(&c, http.Clients, serveMux)

	if c.ServerPort != "" && c.ServerIp != "" {
		go http.StartWebserver(&c)
	}

	return &LogdyInstance{
		config: &config,
	}
}
