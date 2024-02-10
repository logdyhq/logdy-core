package main

import (
	"io"
	"os"

	"github.com/nxadm/tail"
	"github.com/sirupsen/logrus"
)

func followFiles(ch chan Message, files []string) {

	for _, file := range files {

		_, err := os.Stat(file)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"path":  file,
				"error": err.Error(),
			}).Error("Following file changes failed")
			continue
		} else {
			logger.WithFields(logrus.Fields{
				"path": file,
			}).Info("Following file changes")
		}

		go func(file string) {
			t, err := tail.TailFile(
				file, tail.Config{Follow: true, ReOpen: true, Location: &tail.SeekInfo{Offset: 0, Whence: io.SeekEnd}})
			if err != nil {
				logger.WithFields(logrus.Fields{
					"path":  file,
					"error": err.Error(),
				}).Error("Following file changes failed")
			}

			for line := range t.Lines {
				produce(ch, line.Text, MessageTypeStdout, &MessageOrigin{File: file})
			}

		}(file)
	}

}
