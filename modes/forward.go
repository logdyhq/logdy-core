package modes

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/logdyhq/logdy-core/models"
	"github.com/logdyhq/logdy-core/utils"
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
		input, err := readFullLine(reader)
		utils.Logger.WithField("line", string(input)).Debug("Stdin line received")
		if err != nil && err != io.EOF {
			utils.Logger.Error("could not process input", err)
			return
		}

		if len(input) != 0 {
			fmt.Fprint(connClient, string(input)+"\n")
		}

		if err == io.EOF {
			return
		}
	}
}

func ConsumeStdin(ch chan models.Message) {

	reader := bufio.NewReader(os.Stdin)
	for {
		input, err := readFullLine(reader)
		utils.Logger.WithField("line", string(input)).Debug("Stdin line received")

		if err == io.EOF {
			utils.Logger.Debug("EOF")
			return
		}

		if err != nil {
			utils.Logger.Error("could not process input in stdin", err)
			return
		}

		ProduceMessageString(ch, string(input), models.MessageTypeStdout, nil)
	}
}

func readFullLine(reader *bufio.Reader) ([]byte, error) {
	var input []byte
	moreInput, isPrefix, err := reader.ReadLine()
	input = append(input, moreInput...)
	if err != nil {
		return input, err
	}
	for isPrefix {
		var moreInput []byte
		moreInput, isPrefix, err = reader.ReadLine()
		input = append(input, moreInput...)
		if err != nil {
			return input, err
		}
	}
	return input, err
}
