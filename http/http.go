package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/logdyhq/logdy-core/utils"

	"github.com/sirupsen/logrus"
)

func getClientId(r *http.Request) (string, error) {
	kname := "logdy-client-id"
	cid := r.Header.Get(kname)

	if cid == "" {
		cid = r.URL.Query().Get(kname)
	}

	if cid == "" {
		return "", errors.New("missing client id")
	}

	return cid, nil
}

func getClientOrErr(r *http.Request, w http.ResponseWriter, clients *ClientsStruct) *Client {
	cid, err := getClientId(r)

	if err != nil {
		utils.Logger.Error("Missing client id")
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	cl, ok := clients.GetClient(cid)

	if !ok {
		utils.Logger.WithField("client_id", cid).Error("Missing client")
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	return cl
}

func httpError(err string, w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": err,
	})
}

func normalizeHttpPathPrefix(config *Config) {
	if len(config.HttpPathPrefix) > 0 && config.HttpPathPrefix[0] != byte('/') {
		config.HttpPathPrefix = "/" + config.HttpPathPrefix
	}

	if len(config.HttpPathPrefix) == 0 {
		config.HttpPathPrefix = "/"
	}

	if strings.LastIndex(config.HttpPathPrefix, "/") != len(config.HttpPathPrefix)-1 {
		config.HttpPathPrefix = config.HttpPathPrefix + "/"
	}
}

type hand interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	Handle(pattern string, handler http.Handler)
}

func HandleHttp(config *Config, clients *ClientsStruct, serveMux hand) {

	assets, _ := Assets()

	if config.BulkWindowMs > 0 {
		BULK_WINDOW_MS = config.BulkWindowMs
	} else {
		BULK_WINDOW_MS = 100
	}

	normalizeHttpPathPrefix(config)

	// Use the file system to serve static files
	fs := http.FileServer(http.FS(assets))

	v := reflect.ValueOf(serveMux)
	if serveMux == nil || v.IsNil() {
		utils.Logger.Debug("Using net/http")
		http.Handle(config.HttpPathPrefix, http.StripPrefix(config.HttpPathPrefix, fs))
		http.HandleFunc(config.HttpPathPrefix+"api/check-pass", handleCheckPass(config.UiPass))
		http.HandleFunc(config.HttpPathPrefix+"api/status", handleStatus(config))
		http.HandleFunc(config.HttpPathPrefix+"api/client/set-status", handleClientStatus(clients))
		http.HandleFunc(config.HttpPathPrefix+"api/client/load", handleClientLoad(clients))
		http.HandleFunc(config.HttpPathPrefix+"api/client/peek-log", handleClientPeek(clients))
		http.HandleFunc(config.HttpPathPrefix+"ws", handleWs(config.UiPass, clients))

		http.HandleFunc(config.HttpPathPrefix+"api/log", apiKeyMiddleware(config.ApiKey, handleLog(Ch)))
	} else {
		utils.Logger.Debug("Using serveMux", serveMux)
		serveMux.Handle(config.HttpPathPrefix, http.StripPrefix(config.HttpPathPrefix, fs))
		serveMux.HandleFunc(config.HttpPathPrefix+"api/check-pass", handleCheckPass(config.UiPass))
		serveMux.HandleFunc(config.HttpPathPrefix+"api/status", handleStatus(config))
		serveMux.HandleFunc(config.HttpPathPrefix+"api/client/set-status", handleClientStatus(clients))
		serveMux.HandleFunc(config.HttpPathPrefix+"api/client/load", handleClientLoad(clients))
		serveMux.HandleFunc(config.HttpPathPrefix+"api/client/peek-log", handleClientPeek(clients))
		serveMux.HandleFunc(config.HttpPathPrefix+"ws", handleWs(config.UiPass, clients))

		serveMux.HandleFunc(config.HttpPathPrefix+"api/log", apiKeyMiddleware(config.ApiKey, handleLog(Ch)))
	}

}

func StartWebserver(config *Config) {
	utils.Logger.Debug("Starting webserver")
	utils.Logger.WithFields(logrus.Fields{
		"port": config.ServerPort,
	}).Info("WebUI started, visit http://" + config.ServerIp + ":" + config.ServerPort + config.HttpPathPrefix)

	err := http.ListenAndServe(config.ServerIp+":"+config.ServerPort, nil)

	if err != nil {
		panic(err)
	}
}

type Config struct {
	AnalyticsDisabled bool
	UiPass            string
	ConfigFilePath    string
	BulkWindowMs      int64
	HttpPathPrefix    string
	ApiKey            string

	ServerPort string
	ServerIp   string

	AppendToFile    string
	AppendToFileRaw bool
	MaxMessageCount int64

	LogLevel       utils.LOG_LEVEL
	LogInterceptor utils.LogInterceptor
}
