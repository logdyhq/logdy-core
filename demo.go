package main

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

func generateRandomData(jsonFormat bool, numPerSec int, ch chan Message) {

	if numPerSec > 100 {
		numPerSec = 100
	}

	if numPerSec <= 0 {
		numPerSec = 1
	}

	for {
		var msg string
		if jsonFormat {
			msg = generateJsonRandomData()
		} else {
			msg = generateTextRandomData()
		}
		produce(ch, msg, MessageTypeStdout)
		time.Sleep(time.Duration((1 / float64(numPerSec)) * float64(time.Second)))
	}

}

func generateTextRandomData() string {
	return strings.Join([]string{
		gofakeit.UUID(),
		gofakeit.DomainName(),
		gofakeit.IPv4Address(),
		gofakeit.URL(),
		gofakeit.LogLevel("log"),
		gofakeit.UserAgent(),
		gofakeit.HTTPMethod(),
	}, " | ")
}

func generateJsonRandomData() string {
	val, _ := json.Marshal(map[string]string{
		"uuid":   gofakeit.UUID(),
		"domain": gofakeit.DomainName(),
		"ipv4":   gofakeit.IPv4Address(),
		"url":    gofakeit.URL(),
		"level":  gofakeit.LogLevel("log"),
		"ua":     gofakeit.UserAgent(),
		"method": gofakeit.HTTPMethod(),
	})

	return string(val)
}
