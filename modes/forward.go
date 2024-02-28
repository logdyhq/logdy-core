package modes

import (
	"bufio"
	"fmt"
	"io"
	"logdy/utils"
	"net"
	"os"

	"logdy/models"
)

func ConsumeStdinAndForwardToPort(ip string, port string) {
	utils.Logger.WithField("port", port).Info("Accept stdin and forward to port")
	connClient, err := net.Dial("tcp", ip+":"+port)

	if err != nil {
		utils.Logger.WithField("error", err).Error("Error while connecting to port")
		panic(err)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		input, _, err := reader.ReadLine()
		utils.Logger.WithField("line", string(input)).Debug("Stdin line received")
		if err != nil {
			utils.Logger.Error("could not process input", err)
			return
		}

		fmt.Fprint(connClient, string(input)+"\n")
	}
}

func ConsumeStdin(ch chan models.Message) {

	reader := bufio.NewReader(os.Stdin)
	for {
		input, _, err := reader.ReadLine()
		utils.Logger.WithField("line", string(input)).Debug("Stdin line received")

		if err == io.EOF {
			utils.Logger.Debug("EOF")
			return
		}

		if err != nil {
			utils.Logger.Error("could not process input in stdin", err)
			return
		}

		produce(ch, string(input), models.MessageTypeStdout, nil)
	}
}
