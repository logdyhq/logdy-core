package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateTextRandomData(t *testing.T) {
	assert.Equal(t, len(generateTextRandomData()) > 10, true)
}
func TestGenerateJsonRandomData(t *testing.T) {
	assert.Equal(t, len(generateJsonRandomData()) > 10, true)
}

func TestGenerateRandomData(t *testing.T) {
	ch := make(chan Message, 100)

	go generateRandomData(true, 100, ch)

	i := 0
	var firstM time.Time
	var lastM time.Time
	var msg Message
	for {
		msg = <-ch
		i++
		if i == 1 {
			firstM = msg.Ts
		}

		if i == 10 {
			lastM = msg.Ts
			break
		}
	}

	assert.Equal(t, msg.IsJson, true)
	assert.Greater(t, lastM.Sub(firstM).Milliseconds(), int64(90))
}
func TestGenerateRandomData2(t *testing.T) {
	ch := make(chan Message, 100)

	go generateRandomData(false, 100, ch)

	i := 0
	var firstM time.Time
	var lastM time.Time
	var msg Message
	for {
		msg = <-ch
		i++
		if i == 1 {
			firstM = msg.Ts
		}

		if i == 10 {
			lastM = msg.Ts
			break
		}
	}

	assert.Equal(t, msg.IsJson, true)
	assert.Greater(t, lastM.Sub(firstM).Milliseconds(), int64(90))
}
