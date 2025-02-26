package tests

import (
	"bufio"
	"io"
	"net"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLogdyE2E_Forward(t *testing.T) {
	// Track received messages
	msgReceived := []string{}

	// Setup wait groups for coordination
	wg := sync.WaitGroup{}
	wgServer := sync.WaitGroup{}
	wg.Add(3) // Expect 3 messages
	wgServer.Add(1)

	// Start TCP server
	go func() {
		l, err := net.Listen("tcp", ":8475")
		if err != nil {
			panic(err)
		}
		defer l.Close()
		wgServer.Done()

		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			msgReceived = append(msgReceived, scanner.Text())
			wg.Done()
		}
	}()

	// Wait for server to be ready
	wgServer.Wait()

	// Start logdy process
	cmd := exec.Command("go", "run", "../.", "forward", "8475")

	// Get stdin pipe
	stdin, err := cmd.StdinPipe()
	assert.NoError(t, err)

	// Get stdout pipe for logging
	stdout, err := cmd.StdoutPipe()
	assert.NoError(t, err)

	// Start reading stdout in background
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			// Just consume stdout to prevent blocking
			_ = scanner.Text()
		}
	}()

	// Start the process
	err = cmd.Start()
	assert.NoError(t, err)

	// Give the process a moment to start up
	time.Sleep(1 * time.Second)

	// Write test data to stdin
	testLines := []string{
		"test line 1",
		"test line 2",
		"test line 3",
	}

	for _, line := range testLines {
		_, err := io.WriteString(stdin, line+"\n")
		assert.NoError(t, err)
	}

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - all messages received
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for messages")
	}

	// Kill the process
	if err := cmd.Process.Kill(); err != nil {
		t.Errorf("Failed to kill process: %v", err)
	}
	cmd.Wait()
	// Verify received messages
	assert.Equal(t, len(testLines), len(msgReceived))
	for i, testLine := range testLines {
		assert.Equal(t, testLine, msgReceived[i])
	}
}
