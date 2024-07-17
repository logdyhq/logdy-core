package modes

import (
	"bufio"
	"net"
	"os"
	"strings"
	"sync"
	"testing"

	. "github.com/logdyhq/logdy-core/models"
	"github.com/stretchr/testify/assert"
)

func TestConsumeStdinAndForwardToPort(t *testing.T) {

	msgReceived := []string{}
	wg := sync.WaitGroup{}
	wgServer := sync.WaitGroup{}
	wg.Add(2)
	wgServer.Add(1)
	go func() {
		l, err := net.Listen("tcp", ":8123")
		if err != nil {
			panic(err)
		}
		defer l.Close()
		wgServer.Done()
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			defer conn.Close()

			scanner := bufio.NewScanner(conn)

			for scanner.Scan() {
				msgReceived = append(msgReceived, scanner.Text())
				wg.Done()
			}

		}
	}()

	userInput := "lineA\nlineB\n"

	funcDefer, err := mockStdin(t, userInput)

	if err != nil {
		t.Fatal(err)
	}

	defer funcDefer()

	wgServer.Wait()
	ConsumeStdinAndForwardToPort("", "8123")

	wg.Wait()

	assert.Equal(t, len(msgReceived), 2)
	assert.Equal(t, msgReceived[0], "lineA")
	assert.Equal(t, msgReceived[1], "lineB")
}

func TestConsumeStdin(t *testing.T) {
	userInput := "line1\nline2"

	funcDefer, err := mockStdin(t, userInput)
	if err != nil {
		t.Fatal(err)
	}

	defer funcDefer()

	ch := make(chan Message, 10)
	go ConsumeStdin(ch)

	i := 0
	for {
		i++
		msg := <-ch
		if i == 1 {
			assert.Equal(t, "line1", msg.Content)
		}
		if i == 2 {
			assert.Equal(t, "line2", msg.Content)
			break
		}
	}

	assert.Equal(t, 2, i)
}

func TestConsumeStdinLong(t *testing.T) {
	userInput := strings.Repeat("a", 10000)

	funcDefer, err := mockStdin(t, userInput)
	if err != nil {
		t.Fatal(err)
	}

	defer funcDefer()

	ch := make(chan Message, 10)
	go ConsumeStdin(ch)

	msg := <-ch
	if !assert.Equal(t, len(userInput), len(msg.Content)) {
		return
	}
	assert.Equal(t, userInput, msg.Content)
}

func mockStdin(t *testing.T, dummyInput string) (funcDefer func(), err error) {
	t.Helper()

	oldOsStdin := os.Stdin

	tmpfile, err := os.CreateTemp(t.TempDir(), t.Name())
	if err != nil {
		return nil, err
	}

	content := []byte(dummyInput)

	if _, err := tmpfile.Write(content); err != nil {
		return nil, err
	}

	if _, err := tmpfile.Seek(0, 0); err != nil {
		return nil, err
	}

	// Set stdin to the temp file
	os.Stdin = tmpfile

	return func() {
		// clean up
		os.Stdin = oldOsStdin
		os.Remove(tmpfile.Name())
	}, nil
}
