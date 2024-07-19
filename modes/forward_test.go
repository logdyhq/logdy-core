package modes

import (
	"bufio"
	"io"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"

	. "github.com/logdyhq/logdy-core/models"
	"github.com/logdyhq/logdy-core/utils"
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

func TestConsumeStdinAndForwardToPortLong(t *testing.T) {
	msgReceived := []string{}
	wg := sync.WaitGroup{}
	wgServer := sync.WaitGroup{}
	wg.Add(1)
	wgServer.Add(1)
	go func() {
		l, err := net.Listen("tcp", ":8124")
		if err != nil {
			panic(err)
		}
		defer l.Close()
		wgServer.Done()

		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		scanner := bufio.NewScanner(conn)
		scanner.Scan()
		msgReceived = append(msgReceived, scanner.Text())
		wg.Done()
	}()

	userInput := "b" + strings.Repeat("a", 9999)

	funcDefer, err := mockStdin(t, userInput)

	if err != nil {
		t.Fatal(err)
	}

	defer funcDefer()

	wgServer.Wait()
	ConsumeStdinAndForwardToPort("", "8124")

	wg.Wait()

	assert.Equal(t, len(msgReceived), 1)
	if !assert.Equal(t, len(msgReceived[0]), len(userInput)) {
		return
	}
	assert.Equal(t, userInput, msgReceived[0])
}

func TestConsumeStdinAndForwardToPortEofLine(t *testing.T) {
	stubLoggerOut := StubWriter{
		outputs1: []int{0},
		outputs2: []error{nil},
	}
	oldLoggerOut := utils.Logger.Out
	utils.Logger.Out = &stubLoggerOut
	defer func() { utils.Logger.Out = oldLoggerOut }()

	msgReceived := ""
	wg := sync.WaitGroup{}
	wgServer := sync.WaitGroup{}
	wg.Add(1)
	wgServer.Add(1)
	go func() {
		l, err := net.Listen("tcp", ":8124")
		if err != nil {
			panic(err)
		}
		defer l.Close()
		wgServer.Done()

		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		scanner := bufio.NewScanner(conn)
		scanner.Scan()
		msgReceived = scanner.Text()
		wg.Done()
	}()

	userInput := "11111"

	funcStdinDefer, err := mockStdin(t, userInput)
	if err != nil {
		t.Fatal(err)
	}
	defer funcStdinDefer()

	wgServer.Wait()
	ConsumeStdinAndForwardToPort("", "8124")
	wg.Wait()

	if !assert.Equal(t, userInput, msgReceived) {
		return
	}
	messages := stubLoggerOut.InputFieldsAsStrings(t, "msg")
	assert.Equal(t, []string{`"Accept stdin and forward to port"`}, messages)
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
	userInput := "b" + strings.Repeat("a", 9999)

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

type StubWriter struct {
	inputs      [][]byte
	outputs1    []int
	outputs2    []error
	timesCalled int
}

var _ io.Writer = (*StubWriter)(nil)

func (s *StubWriter) Write(b []byte) (n int, err error) {
	s.inputs = append(s.inputs, b)
	idx := min(len(s.outputs1)-1, len(s.outputs2)-1, s.timesCalled)
	n = s.outputs1[idx]
	err = s.outputs2[idx]
	s.timesCalled++
	return n, err
}

func (s *StubWriter) InputFieldsAsStrings(t *testing.T, fieldName string) []string {
	t.Helper()
	var messages []string
	for _, v := range s.inputs {
		message := loggingValue(t, string(v), fieldName)
		messages = append(messages, message)
	}
	return messages
}

func loggingValue(t *testing.T, logEntry string, field string) string {
	t.Helper()
	if logEntry[len(logEntry)-1] == '\n' {
		logEntry = logEntry[0 : len(logEntry)-1]
	}
	r, err := regexp.Compile(`[a-zA-Z]+=(?:[a-zA-Z0-9]+|(?:\"[\w\s:\.]+\"))`)
	if err != nil {
		t.Fatalf("failed to compile regex: %v", err)
	}
	entries := r.FindAll(([]byte)(logEntry), -1)
	for _, v := range entries {
		equalsIndex := strings.Index(string(v), "=")
		key := string(v[0:equalsIndex])
		value := string(v[equalsIndex+1:])
		if key == field {
			return value
		}
	}
	t.Fatalf("expected logging field to have key %v but was not found", field)
	return ""
}
