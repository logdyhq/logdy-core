package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
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

func handleWs(uiPass string, msgs <-chan Message, clients *Clients) func(w http.ResponseWriter, r *http.Request) {

	wsUpgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
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

		ch := clients.Join(1000, true)
		clientId := ch.id

		bts, err := json.Marshal(ClientJoined{
			BaseMessage: BaseMessage{
				MessageType: MessageTypeClientJoined,
			},
			ClientId: ch.id,
		})

		if err != nil {
			fmt.Println(err)
			return
		}

		err = conn.WriteMessage(1, bts)

		if err != nil {
			fmt.Println(err)
			return
		}

		go func(clienId string) {
			for {
				time.Sleep(1 * time.Second)
				_, _, err := conn.ReadMessage()
				if err != nil {
					logger.Debug(err)
					logger.WithField("client_id", clientId).Info("Closed client")
					clients.Close(clientId)
					return
				}
			}
		}(clientId)

		for {
			msgs := <-ch.ch
			bulk := MessageBulk{
				BaseMessage: BaseMessage{
					MessageType: MessageTypeLogBulk,
				},
				Messages: msgs,
				Status:   clients.Stats(),
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

func getClientOrErr(r *http.Request, w http.ResponseWriter, clients *Clients) *Client {
	cid, err := getClientId(r)

	if err != nil {
		logger.Error("Missing client id")
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	cl, ok := clients.GetClient(cid)

	if !ok {
		logger.WithField("client_id", cid).Error("Missing client")
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}

	return cl
}

func handleClientStatus(clients *Clients) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		cl := getClientOrErr(r, w, clients)
		if cl == nil {
			return
		}

		status := r.URL.Query().Get("status")

		switch status {
		case string(CURSOR_FOLLOWING):
			clients.ResumeFollowing(cl.id, true)
		case string(CURSOR_STOPPED):
			clients.PauseFollowing(cl.id)
		default:
			logger.Error("Unrecognized status")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func handleClientLoad(clients *Clients) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		cl := getClientOrErr(r, w, clients)
		if cl == nil {
			return
		}

		start := r.URL.Query().Get("start")
		count := r.URL.Query().Get("count")

		startInt, err := strconv.Atoi(start)
		if err != nil {
			logger.Error("Invalid start")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		countInt, err := strconv.Atoi(count)
		if err != nil {
			logger.Error("Invalid count")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		clients.Load(cl.id, startInt, countInt, true)

		w.WriteHeader(http.StatusOK)
	}
}

func handleClientPeek(clients *Clients) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		cl := getClientOrErr(r, w, clients)
		if cl == nil {
			return
		}

		type Req struct {
			Idxs []int `json:"idxs"`
		}

		var p Req

		// Try to decode the request body into the struct. If there is an error,
		// respond to the client with the error message and a 400 status code.
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		msgs := clients.PeekLog(p.Idxs)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(msgs)

	}
}

func handleHttp(msgs <-chan Message, httpPort string, analyticsEnabled bool, uiPass string, configFilePath string, bulkWindowMs int64, maxMessageCount int64) {
	assets, _ := Assets()
	clients := NewClients(msgs, maxMessageCount)

	BULK_WINDOW_MS = bulkWindowMs

	// Use the file system to serve static files
	fs := http.FileServer(http.FS(assets))
	http.Handle("/", http.StripPrefix("/", fs))

	http.HandleFunc("/api/check-pass", handleCheckPass(uiPass))
	http.HandleFunc("/api/status", handleStatus(configFilePath, analyticsEnabled, uiPass))
	http.HandleFunc("/api/client/status", handleClientStatus(clients))
	http.HandleFunc("/api/client/load", handleClientLoad(clients))
	http.HandleFunc("/api/client/peek-log", handleClientPeek(clients))

	// Listen for WebSocket connections on port 8080.
	http.HandleFunc("/ws", handleWs(uiPass, msgs, clients))

	logger.WithFields(logrus.Fields{
		"port": httpPort,
	}).Info("WebUI started, visit http://localhost:" + httpPort)

	http.ListenAndServe(":"+httpPort, nil)
}
