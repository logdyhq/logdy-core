package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClientStartAddToBuffer(t *testing.T) {
	ch := make(chan Message)
	c := NewClients(ch)

	go c.Start()

	assert.Equal(t, len(c.buffer), 0)
	ch <- Message{}
	time.Sleep(1 * time.Millisecond)
	assert.Equal(t, len(c.buffer), 1)
}

func TestClientJoinPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	ch := make(chan Message)
	c := NewClients(ch)
	c.Join(1)
	c.Join(1)
}

func TestClientJoinSingle(t *testing.T) {
	ch := make(chan Message)
	c := NewClients(ch)
	go c.Start()
	client := c.Join(1)

	ch <- Message{Content: "foo"}

	msg := <-client.ch

	assert.Equal(t, 1, len(msg))
	assert.Equal(t, "foo", msg[0].Content)
}

func TestClientJoinSingleAfterMessage(t *testing.T) {
	ch := make(chan Message)
	c := NewClients(ch)
	go c.Start()
	ch <- Message{Content: "foo"}
	client := c.Join(1)

	msg := <-client.ch

	assert.Equal(t, 1, len(msg))
	assert.Equal(t, "foo", msg[0].Content)
}

func TestClientJoinMultiple(t *testing.T) {
	ch := make(chan Message)
	c := NewClients(ch)
	go c.Start()
	client1 := c.Join(1)
	client2 := c.Join(2)
	client3 := c.Join(3)

	ch <- Message{Content: "foo"}

	msg := <-client1.ch
	assert.Equal(t, 1, len(msg))
	assert.Equal(t, "foo", msg[0].Content)

	msg = <-client2.ch
	assert.Equal(t, 1, len(msg))
	assert.Equal(t, "foo", msg[0].Content)

	msg = <-client3.ch
	assert.Equal(t, 1, len(msg))
	assert.Equal(t, "foo", msg[0].Content)
}

func TestClientBulkWindow(t *testing.T) {
	ch := make(chan Message)
	c := NewClients(ch)
	go c.Start()
	client1 := c.Join(1)

	ch <- Message{Content: "foo1"}
	ch <- Message{Content: "foo2"}
	ch <- Message{Content: "foo3"}
	ts := time.Now()
	messages := <-client1.ch

	assert.Equal(t, 3, len(messages))
	assert.GreaterOrEqual(t, int64(time.Since(ts).Milliseconds()), BULK_WINDOW_MS)
	assert.Equal(t, "foo1", messages[0].Content)
	assert.Equal(t, "foo2", messages[1].Content)
	assert.Equal(t, "foo3", messages[2].Content)
}
func TestClientSignalQuit(t *testing.T) {
	ch := make(chan Message)
	c := NewClients(ch)
	go c.Start()

	c.Join(1)
	c.Close(1)
}

func TestClientCloseError(t *testing.T) {
	ch := make(chan Message)
	c := NewClients(ch)
	go c.Start()

	c.Close(1)
	c.Close(1)
}
