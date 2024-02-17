package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

func handleCheckPass(uiPass string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("/api/check-pass")
		pass := r.URL.Query().Get("password")
		if uiPass == "" {
			w.WriteHeader(200)
			return
		}

		if pass == "" || uiPass != pass {
			logger.WithFields(logrus.Fields{
				"ip": r.RemoteAddr,
				"ua": r.Header.Get("user-agent"),
			}).Info("Client denied")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.WriteHeader(200)
	}
}

func handleStatus(configFilePath string, analyticsEnabled bool, uiPass string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		logger.Debug("/api/status")

		configStr := ""
		if configFilePath != "" {
			logger.Debug("Reading config file")
			configStr = loadFile(configFilePath)
		}

		initMsg, _ := json.Marshal(InitMessage{
			BaseMessage: BaseMessage{
				MessageType: MessageTypeInit,
			},
			AnalyticsEnabled: analyticsEnabled,
			AuthRequired:     uiPass != "",
			ConfigStr:        configStr,
		})

		w.Write(initMsg)
	}
}

func handleWs(uiPass string, msgs <-chan Message) func(w http.ResponseWriter, r *http.Request) {
	clients := Clients{
		mu:                 sync.Mutex{},
		mainChan:           msgs,
		clients:            map[int]*Client{},
		currentlyConnected: 0,
		buffer:             []Message{},
	}

	go clients.Start()

	wsUpgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	cid := 0
	return func(w http.ResponseWriter, r *http.Request) {

		if uiPass != "" {
			pass := r.URL.Query().Get("password")
			if pass == "" || uiPass != pass {
				logger.WithFields(logrus.Fields{
					"ip": r.RemoteAddr,
					"ua": r.Header.Get("user-agent"),
				}).Info("Client denied")
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		// Upgrade the HTTP connection to a WebSocket connection.
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		logger.Info("New Web UI client connected")

		cid++
		clientId := cid
		ch := clients.Join(cid)

		go func() {
			for {
				time.Sleep(1 * time.Second)
				_, _, err := conn.ReadMessage()
				if err != nil {
					logger.Debug(err)
					logger.WithField("client_id", clientId).Info("Closed client")
					clients.Close(cid)
					return
				}
			}
		}()

		for {
			msgs := <-ch.ch
			bulk := MessageBulk{
				BaseMessage: BaseMessage{
					MessageType: MessageTypeLogBulk,
				},
				Messages: msgs,
			}

			logger.WithField("count", len(msgs)).Debug("Received messages")

			if logger.Level <= logrus.DebugLevel {
				for _, msg := range msgs {
					mbts, _ := json.Marshal(msg)
					logger.WithFields(logrus.Fields{
						"msg":      trunc(string(mbts), 45),
						"clientId": clientId,
					}).Debug("Sending message through WebSocket")
				}
			}

			bts, err := json.Marshal(bulk)
			if err != nil {
				fmt.Println("Error while serializing message", err)
				continue
			}

			err = conn.WriteMessage(1, bts)

			if err != nil {
				logger.Error("Err", err)
				clients.Close(clientId)
				logger.WithField("client_id", clientId).Info("Closed client")
				break
			}
		}

	}
}

func handleHttp(msgs <-chan Message, httpPort string, analyticsEnabled bool, uiPass string, configFilePath string) {
	assets, _ := Assets()

	// Use the file system to serve static files
	fs := http.FileServer(http.FS(assets))
	http.Handle("/", http.StripPrefix("/", fs))

	http.HandleFunc("/api/check-pass", handleCheckPass(uiPass))
	http.HandleFunc("/api/status", handleStatus(configFilePath, analyticsEnabled, uiPass))

	// Listen for WebSocket connections on port 8080.
	http.HandleFunc("/ws", handleWs(uiPass, msgs))

	logger.WithFields(logrus.Fields{
		"port": httpPort,
	}).Info("WebUI started, visit http://localhost:" + httpPort)

	http.ListenAndServe(":"+httpPort, nil)
}
