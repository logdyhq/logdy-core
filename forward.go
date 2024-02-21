package main

import (
	"bufio"
	"fmt"
	"io"
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
			logger.Error("could not process input", err)
			return
		}

		fmt.Fprint(connClient, string(input)+"\n")
	}
}

func consumeStdin(ch chan Message) {

	reader := bufio.NewReader(os.Stdin)
	for {
		input, _, err := reader.ReadLine()
		logger.WithField("line", string(input)).Debug("Stdin line received")

		if err == io.EOF {
			logger.Debug("EOF")
			return
		}

		if err != nil {
			logger.Error("could not process input in stdin", err)
			return
		}

		produce(ch, string(input), MessageTypeStdout, nil)
	}
}
