package utils

import (
	"encoding/json"
	"io"
	"os"

	"github.com/logdyhq/logdy-core/models"
)

func ProcessIncomingMessages(ch chan models.Message, appendToFile string, appendToFileRaw bool) chan models.Message {
	mainChan := make(chan models.Message, 1000)

	if appendToFile != "" {
		go func() {
			f, err := os.OpenFile(appendToFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			Logger.WithField("file", appendToFile).Info("Writing messages to a file")
			if appendToFileRaw {
				Logger.WithField("file", appendToFile).Info("Appending raw messages")
			}
			rewriteMessagesToWriter(ch, mainChan, f, appendToFileRaw)
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
