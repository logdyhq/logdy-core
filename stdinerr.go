package main

import (
	"bufio"
	"io"
	"os/exec"
)

func readOutput(reader io.Reader, outputCh chan Message, messageType LogType) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		produce(outputCh, scanner.Text(), messageType)
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
