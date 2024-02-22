package main

import (
	"io"
	"math/rand"
	"os"
	"time"
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

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
