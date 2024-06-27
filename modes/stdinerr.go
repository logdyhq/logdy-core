package modes

import (
	"bufio"
	"io"
	"os/exec"

	"github.com/logdyhq/logdy-core/models"
	"github.com/logdyhq/logdy-core/utils"
)

func readOutput(reader io.Reader, outputCh chan models.Message, messageType models.LogType) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		ProduceMessageString(outputCh, scanner.Text(), messageType, nil)
	}
}

func StartCmd(ch chan models.Message, cmdStr string, args []string) {
	cmd := exec.Command(cmdStr, args...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		utils.Logger.Error("Error creating stdout pipe:", err)
		return
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		utils.Logger.Error("Error creating stderr pipe:", err)
		return
	}

	if err := cmd.Start(); err != nil {
		utils.Logger.Error("Error starting command:", err)
		return
	}
	go readOutput(stdoutPipe, ch, models.MessageTypeStdout)
	go readOutput(stderrPipe, ch, models.MessageTypeStderr)
	utils.Logger.Info("Listening to stdout/stderr")
}
