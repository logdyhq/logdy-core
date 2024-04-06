package utils

import (
	"io"
	"math/rand"
	"os"
	"time"
)

func LoadFile(configFilePath string) string {
	f, err := os.OpenFile(configFilePath, os.O_RDONLY, 0644)

	if err != nil {
		Logger.Error("Error while loading config file")
		panic(err)
	}

	bytes, err := io.ReadAll(f)

	if err != nil {
		Logger.Error("Error while loading config file")
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

func Trunc(str string, limit int) string {

	if len(str) <= limit {
		return str
	}

	return str[:limit] + "..."
}

func PickRandom[T any](slice []T) T {
	if len(slice) == 0 {
		panic("slice is empty")
	}
	index := rand.Intn(len(slice))
	return slice[index]
}
