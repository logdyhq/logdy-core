package modes

import (
	"bufio"
	"net"
	"os"

	"github.com/logdyhq/logdy-core/models"
	"github.com/logdyhq/logdy-core/utils"
	"github.com/sirupsen/logrus"
)

func handleConnection(conn net.Conn, ch chan models.Message, port string) {
	defer conn.Close()

	// Create a new bufio.Scanner to read lines from the connection
	scanner := bufio.NewScanner(conn)

	// Read lines from the connection and write them to the channel
	for scanner.Scan() {
		ProduceMessageString(ch, scanner.Text(), models.MessageTypeStdout, &models.MessageOrigin{Port: port, File: ""})
	}
}

func StartSocketServers(ch chan models.Message, ip string, ports []string) {
	for _, port := range ports {
		go startSocketServer(ch, ip, port)
	}
}

func startSocketServer(ch chan models.Message, ip string, port string) {

	addr := ip + ":" + port
	// Start the TCP server
	server, err := net.Listen("tcp", addr)
	if err != nil {
		utils.Logger.Error("Error starting server:", err)
		os.Exit(1)
	}
	defer server.Close()

	utils.Logger.WithFields(logrus.Fields{
		"address": addr,
	}).Info("TCP Server is listening")

	// Accept incoming connections and handle them in a separate goroutine
	for {
		conn, err := server.Accept()
		if err != nil {
			utils.Logger.Error("Error accepting connection:", err)
			continue
		}

		utils.Logger.Info("Connection accepted")
		go handleConnection(conn, ch, port)
	}
}
