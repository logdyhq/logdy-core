package main

import (
	"bufio"
	"encoding/json"
	"io"
	"os/exec"

	"github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
)

func readOutput(reader io.Reader, outputCh chan<- Message, messageType MessageType) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		logger.WithFields(logrus.Fields{
			"msg": scanner.Text(),
		}).Debug("Got messages from process STDOUT/STDERR")
		validJson := fastjson.Validate(scanner.Text())
		var cs json.RawMessage
		if validJson == nil {
			cs = json.RawMessage(scanner.Bytes())
		}

		outputCh <- Message{Mtype: messageType, Content: scanner.Text(), JsonContent: cs, IsJson: validJson == nil}
	}
}

func startCmd(ch chan Message, cmdStr string, args []string) {
	cmd := exec.Command(cmdStr, args...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		logger.Error("Error creating stdout pipe:", err)
		return
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		logger.Error("Error creating stderr pipe:", err)
		return
	}

	if err := cmd.Start(); err != nil {
		logger.Error("Error starting command:", err)
		return
	}
	go readOutput(stdoutPipe, ch, MessageTypeStdout)
	go readOutput(stderrPipe, ch, MessageTypeStderr)
	logger.Info("Listening to stdout/stderr")
}
