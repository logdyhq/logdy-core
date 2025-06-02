package utils

import (
	"encoding/json"
	"io"
	"os"

	"github.com/logdyhq/logdy-core/models"
)

func ProcessIncomingMessages(ch chan models.Message, appendToFile string, appendToFileRaw bool) chan models.Message {
	return ProcessIncomingMessagesWithRotation(ch, appendToFile, appendToFileRaw, 0, 0)
}

// ProcessIncomingMessagesWithRotation processes incoming messages with optional log rotation
// maxSize: maximum size of the log file in bytes before rotation (0 = no rotation)
func ProcessIncomingMessagesWithRotation(ch chan models.Message, appendToFile string, appendToFileRaw bool, maxSize int64, maxBackups int) chan models.Message {
	mainChan := make(chan models.Message, 1000)

	if appendToFile != "" {
		go func() {
			var writer io.WriteCloser
			var err error

			if maxSize > 0 {
				// Use rotating writer
				writer, err = NewRotatingWriter(appendToFile, maxSize, maxBackups)
				if err != nil {
					panic(err)
				}
				Logger.WithField("file", appendToFile).WithField("maxSize", maxSize).WithField("maxBackups", maxBackups).Debug("Writing messages to a rotating file")
			} else {
				// Use simple file writer
				writer, err = os.OpenFile(appendToFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					panic(err)
				}
				Logger.WithField("file", appendToFile).Debug("Writing messages to a file")
			}
			defer writer.Close()

			if appendToFileRaw {
				Logger.WithField("file", appendToFile).Debug("Appending raw messages")
			}
			rewriteMessagesToWriter(ch, mainChan, writer, appendToFileRaw)
		}()
	} else {
		go rewriteMessages(ch, mainChan)
	}

	return mainChan
}

func rewriteMessagesToWriter(source chan models.Message, dest chan models.Message, f io.Writer, raw bool) {
	for {
		msg := <-source
		toSerialize := msg
		var bts []byte
		if raw {
			bts = []byte(msg.Content)
		} else {
			bts, _ = json.Marshal(toSerialize)

		}
		bts = append(bts, []byte("\n")...)
		if _, err := f.Write(bts); err != nil {
			panic(err)
		}
		dest <- msg
	}
}

func rewriteMessages(source chan models.Message, dest chan models.Message) {
	for {
		msg := <-source
		dest <- msg
	}
}
