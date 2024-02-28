package modes

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"logdy/models"
)

func TestFollowFiles(t *testing.T) {

	ch := make(chan models.Message)
	ctx, cancel := context.WithCancel(context.Background())

	f, err := os.CreateTemp("", "sample")

	go func(ctx context.Context) {
		i := 0
		for {
			if ctx.Err() != nil {
				return
			}

			f.Write([]byte("foobar" + strconv.Itoa(i)))
			time.Sleep(1 * time.Millisecond)
			i++
		}
	}(ctx)

	assert.Equal(t, err, nil)

	go FollowFiles(ch, []string{f.Name()})

	received := 0
	for received < 20 {
		<-ch
		received++

	}
	cancel()

	assert.GreaterOrEqual(t, received, 20)

}
