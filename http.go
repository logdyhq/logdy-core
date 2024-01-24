package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type MessageType int

const MessageTypeStdout MessageType = 1
const MessageTypeStderr MessageType = 2

type Message struct {
	Mtype       MessageType     `json:"message_type"`
	Content     string          `json:"content"`
	JsonContent json.RawMessage `json:"json_content"`
	IsJson      bool            `json:"is_json"`
}

type Clients struct {
	mu       sync.Mutex
	mainChan <-chan Message
	clients  map[int]chan Message
}

func (c *Clients) Start() {
	for {
		msg := <-c.mainChan
		c.mu.Lock()
		for _, ch := range c.clients {
			// log.Println("Sending to", cid)
			ch <- msg
		}
		c.mu.Unlock()
	}
}

func (c *Clients) Join(id int) <-chan Message {
	c.clients[id] = make(chan Message, 100)
	return c.clients[id]
}

func (c *Clients) Close(id int) {
	c.mu.Lock()
	delete(c.clients, id)
	c.mu.Unlock()
}

func handleHttp(msgs <-chan Message, httpPort string) {
	// Create a new WebSocket server.
	wsUpgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	clients := Clients{
		mu:       sync.Mutex{},
		mainChan: msgs,
		clients:  map[int]chan Message{},
	}

	go clients.Start()

	cid := 0

	assets, _ := Assets()

	// Use the file system to serve static files
	fs := http.FileServer(http.FS(assets))
	http.Handle("/", http.StripPrefix("/", fs))

	// Listen for WebSocket connections on port 8080.
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// Upgrade the HTTP connection to a WebSocket connection.
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		logger.Info("New client connected")

		cid = cid + 1
		clientId := cid
		ch := clients.Join(cid)

		for {
			msg := <-ch
			bts, err := json.Marshal(msg)

			logger.WithFields(logrus.Fields{
				"msg": string(bts),
			}).Debug("Sending message through WebSocket")

			if err != nil {
				fmt.Printf("Received message %+v", msg)
				fmt.Println("Error while serializing message", err)
				continue
			}

			err = conn.WriteMessage(1, bts)

			if err != nil {
				log.Println("Err", err)
				clients.Close(clientId)
				log.Println("Closed client")
				break
			}
		}

	})

	logger.WithFields(logrus.Fields{
		"port": httpPort,
	}).Info("WebUI started, visit http://localhost:" + httpPort)

	http.ListenAndServe(":"+httpPort, nil)
}
