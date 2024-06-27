package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	logdy "github.com/logdyhq/logdy-core/logdy"
)

func main() {

	switch os.Args[1] {
	case "with-webserver":
		// go run main.go with-webserver
		exampleWithWebserver()
	case "with-serve-mux":
		// go run main.go with-serve-mux
		exampleWithServeMux()
	case "without-webserver":
		// go run main.go without-webserver
		exampleWithoutWebserver()
	}
}

func exampleWithoutWebserver() {
	logdyLogger := logdy.InitializeLogdy(logdy.Config{
		ServerIp:       "127.0.0.1",
		ServerPort:     "8080",
		UiPass:         "foobar12345",
		HttpPathPrefix: "_logdy-ui",
		LogLevel:       logdy.LOG_LEVEL_NORMAL,
		LogInterceptor: func(entry *logdy.LogEntry) {
			log.Println("Logdy internal log message intercepted", entry.Message, entry.Data, entry.Time)
		},
	}, nil)

	go func() {
		for {
			logdyLogger.Log(logdy.Fields{
				"foo":  "bar",
				"time": time.Now(),
			})
			logdyLogger.LogString("This is just a string " + time.Now().String())
			time.Sleep(1000 * time.Millisecond)
		}
	}()

	<-context.Background().Done()
}

func exampleWithWebserver() {
	var logger logdy.Logdy

	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		logger.Log(logdy.Fields{
			"url": r.URL.String(),
			"ua":  r.Header.Get("user-agent"),
		})
		fmt.Fprintf(w, "Hello world!")
	})

	logger = logdy.InitializeLogdy(logdy.Config{
		HttpPathPrefix: "/_logdy-ui",
		LogLevel:       logdy.LOG_LEVEL_VERBOSE,
	}, nil)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Logger is a middleware handler that does request logging
type Logger struct {
	logdy   logdy.Logdy
	handler http.Handler
}

// ServeHTTP handles the request by passing it to the real
// handler and logging the request details
func (l *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	l.handler.ServeHTTP(w, r)

	// If this is a request to Logdy backend, ignore it
	if strings.HasPrefix(r.URL.Path, l.logdy.Config().HttpPathPrefix) {
		return
	}

	l.logdy.Log(logdy.Fields{
		"ua":     r.Header.Get("user-agent"),
		"method": r.Method,
		"path":   r.URL.Path,
		"query":  r.URL.RawQuery,
		"time":   time.Since(start),
	})
}

func exampleWithServeMux() {

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})
	mux.HandleFunc("/v1/time", func(w http.ResponseWriter, r *http.Request) {
		curTime := time.Now().Format(time.Kitchen)
		w.Write([]byte(fmt.Sprintf("the current time is %v", curTime)))
	})

	logger := logdy.InitializeLogdy(logdy.Config{
		HttpPathPrefix: "/_logdy-ui",
		LogLevel:       logdy.LOG_LEVEL_NORMAL,
	}, mux)

	addr := ":8082"
	log.Printf("server is listening at %s", addr)
	log.Fatal(http.ListenAndServe(addr, &Logger{logdy: logger, handler: mux}))
}
