package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func consumeStdinAndForwardToPort(ip string, port string) {
	logger.WithField("port", port).Info("Accept stdin and forward to port")
	connClient, err := net.Dial("tcp", ip+":"+port)

	if err != nil {
		logger.WithField("error", err).Error("Error while connecting to port")
		panic(err)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		input, _, err := reader.ReadLine()
		logger.WithField("line", string(input)).Debug("Stdin line received")
		if err != nil {
			logger.Error("could not process input")
		}

		fmt.Fprint(connClient, string(input)+"\n")
	}
}
