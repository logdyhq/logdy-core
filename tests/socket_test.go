package tests

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLogdyE2E_Socket(t *testing.T) {
	// Channel for collecting messages
	msgChan := make(chan string, 10)

	// Setup wait group for message sending
	var wg sync.WaitGroup
	var wgReady sync.WaitGroup
	wgReady.Add(1)
	wg.Add(3) // Expect 3 messages (1 from each port)

	// Start logdy process with -t flag for stdout output
	cmd := exec.Command("go", "run", "../main.go", "socket", "-t", "8475", "8476", "8477")
	// Get stdout pipe for verifying messages
	stdout, err := cmd.StdoutPipe()
	assert.NoError(t, err)

	// Start reading stdout in background
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, `WebUI started`) {
				wgReady.Done()
			}
			select {
			case msgChan <- line:
				// Message sent to channel
			default:
				// Channel full, ignore additional messages
			}
		}
	}()

	// Start the process
	err = cmd.Start()
	assert.NoError(t, err)

	// Give the process more time to start up and initialize all socket servers
	wgReady.Wait()
	time.Sleep(1 * time.Second)

	// Send test messages to each port
	ports := []string{"8475", "8476", "8477"}
	for _, port := range ports {
		// Try to connect with retries
		var conn net.Conn
		for retries := 0; retries < 3; retries++ {
			conn, err = net.Dial("tcp", fmt.Sprintf("localhost:%s", port))
			if err == nil {
				break
			}
			time.Sleep(1 * time.Millisecond)
		}
		assert.NoError(t, err, "Failed to connect to port %s after retries", port)

		if conn != nil {
			message := fmt.Sprintf("test message on port %s", port)
			_, err = fmt.Fprintln(conn, message)
			assert.NoError(t, err)
			conn.Close()
			wg.Done()
		}
	}

	// Wait with timeout for messages to be sent
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - all messages sent
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for messages to be sent")
	}

	// Collect received messages
	var msgReceived []string
	timeout := time.After(5 * time.Second)

	for len(msgReceived) < 3 {
		select {
		case msg := <-msgChan:
			if strings.Contains(msg, "test message on port") {
				msgReceived = append(msgReceived, msg)
			}
		case <-timeout:
			t.Fatal("Timeout waiting for messages to be received")
		}
	}

	// Kill the process
	if err := cmd.Process.Kill(); err != nil {
		t.Errorf("Failed to kill process: %v", err)
	}

	// Verify we received messages from all ports
	assert.Equal(t, 3, len(msgReceived), "Expected 3 messages, got %d", len(msgReceived))
	for i, port := range ports {
		expectedMsg := fmt.Sprintf("test message on port %s", port)
		assert.Contains(t, msgReceived[i], expectedMsg)
	}
}
