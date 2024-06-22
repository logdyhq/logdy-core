package modes

import (
	"context"
	"testing"

	. "github.com/logdyhq/logdy-core/models"
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
	ctx, cancel := context.WithCancel(context.Background())

	go GenerateRandomData(true, 100, ch, ctx)

	i := 0
	var firstM int64
	var lastM int64
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
	assert.Greater(t, lastM-firstM, int64(90))
	cancel()
}
func TestGenerateRandomData2(t *testing.T) {
	ch := make(chan Message, 100)
	ctx, cancel := context.WithCancel(context.Background())

	go GenerateRandomData(false, 100, ch, ctx)

	i := 0
	var firstM int64
	var lastM int64
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

	assert.Equal(t, msg.IsJson, false)
	assert.Greater(t, lastM-firstM, int64(90))
	cancel()
}
