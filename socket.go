package main

import (
	"bufio"
	"encoding/json"
	"net"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
)

func handleConnection(conn net.Conn, ch chan Message) {
	defer conn.Close()

	// Create a new bufio.Scanner to read lines from the connection
	scanner := bufio.NewScanner(conn)

	// Read lines from the connection and write them to the channel
	for scanner.Scan() {
		line := scanner.Text()
		logger.WithFields(logrus.Fields{
			"msg": line,
		}).Debug("Message received")

		validJson := fastjson.Validate(line)
		var cs json.RawMessage
		if validJson == nil {
			cs = json.RawMessage(line)
		}
		ch <- Message{Mtype: MessageTypeStdout, Content: line, JsonContent: cs, IsJson: validJson == nil}
	}
}

func startSocketServer(ch chan Message, port string) {

	// Start the TCP server
	server, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Error("Error starting server:", err)
		os.Exit(1)
	}
	defer server.Close()

	logger.WithFields(logrus.Fields{
		"port": port,
	}).Info("TCP Server is listening")

	// Accept incoming connections and handle them in a separate goroutine
	for {
		conn, err := server.Accept()
		if err != nil {
			logger.Error("Error accepting connection:", err)
			continue
		}

		logger.Info("Connection accepted")
		go handleConnection(conn, ch)
	}

}
