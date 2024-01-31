package main

import (
	"bufio"
	"net"
	"os"

	"github.com/sirupsen/logrus"
)

func handleConnection(conn net.Conn, ch chan Message) {
	defer conn.Close()

	// Create a new bufio.Scanner to read lines from the connection
	scanner := bufio.NewScanner(conn)

	// Read lines from the connection and write them to the channel
	for scanner.Scan() {
		produce(ch, scanner.Text(), MessageTypeStdout)
	}
}

func startSocketServer(ch chan Message, ip string, port string) {

	addr := ip + ":" + port
	// Start the TCP server
	server, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("Error starting server:", err)
		os.Exit(1)
	}
	defer server.Close()

	logger.WithFields(logrus.Fields{
		"address": addr,
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
