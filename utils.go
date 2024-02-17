package main

import (
	"io"
	"os"
)

func loadFile(configFilePath string) string {
	f, err := os.OpenFile(configFilePath, os.O_RDONLY, 0644)

	if err != nil {
		logger.Error("Error while loading config file")
		panic(err)
	}

	bytes, err := io.ReadAll(f)

	if err != nil {
		logger.Error("Error while loading config file")
		panic(err)
	}

	return string(bytes)
}
