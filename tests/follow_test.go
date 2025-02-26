package tests

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func readOutput(t *testing.T, stdout io.ReadCloser, outputChan chan string, followChan chan bool, wg *sync.WaitGroup) {
	reader := bufio.NewReader(stdout)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			return
		}
		if err != nil {
			t.Logf("Error reading stdout: %v", err)
			return
		}
		// Log all lines for debugging
		t.Logf("Received line: %s", strings.TrimSpace(line))
		// Signal when we see the "Following file changes" message
		if strings.Contains(line, "Following file changes") {
			select {
			case followChan <- true:
			default:
			}
		}
		// Only capture actual test lines, not debug/info messages
		if strings.Contains(line, "test line") && !strings.Contains(line, "level=") {
			outputChan <- strings.TrimSpace(line)
		}
	}
}

func TestLogdyE2E_FollowFullRead(t *testing.T) {
	// Create a named pipe
	pipeName := "/tmp/logdy-test-pipe-full"

	// Remove existing pipe if it exists
	if _, err := os.Stat(pipeName); err == nil {
		os.Remove(pipeName)
	}

	err := exec.Command("mkfifo", pipeName).Run()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	defer os.Remove(pipeName)

	t.Logf("Created named pipe: %s", pipeName)

	// Channel to communicate the pipe writer
	pipeWriterChan := make(chan *os.File)

	// Open pipe for writing in a goroutine
	go func() {
		pipeWriter, err := os.OpenFile(pipeName, os.O_WRONLY, 0644)
		assert.NoError(t, err)
		pipeWriterChan <- pipeWriter // Send the writer to the main goroutine
	}()

	// Start logdy process in follow mode with full-read and fallthrough enabled
	cmd := exec.Command("go", "run", "../.", "follow", "--full-read", "-t", pipeName)

	// Get stdout pipe
	stdout, err := cmd.StdoutPipe()
	assert.NoError(t, err)

	// Create channels to collect output lines and signal following started
	outputChan := make(chan string, 100) // Buffered channel to prevent blocking
	followChan := make(chan bool, 1)     // Channel to signal when following starts

	// Start the process
	err = cmd.Start()
	assert.NoError(t, err)

	// Use WaitGroup to manage the goroutine
	var wg sync.WaitGroup
	wg.Add(1)
	go readOutput(t, stdout, outputChan, followChan, &wg)

	// Wait for the pipe writer
	var pipeWriter *os.File
	select {
	case pipeWriter = <-pipeWriterChan:
		defer pipeWriter.Close()
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for pipe writer")
	}

	// Write all test lines to the pipe
	allTestLines := []string{
		"test line 1",
		"test line 2",
		"test line 3",
		"test line 4",
	}

	for _, line := range allTestLines {
		t.Logf("Writing line: %s", line)
		_, err := pipeWriter.WriteString(line + "\n")
		assert.NoError(t, err)
		// Small delay between writes
		time.Sleep(1 * time.Millisecond)
	}
	wg.Done()

	// Collect output with timeout
	receivedLines := make([]string, 0)
	expectedLines := 4
	timeout := time.After(5 * time.Second)

	for i := 0; i < expectedLines; i++ {
		select {
		case line := <-outputChan:
			t.Logf("Collected line: %s", line)
			receivedLines = append(receivedLines, line)
		case <-timeout:
			t.Fatalf("Timeout waiting for output. Got %d lines, expected %d. Received lines: %v",
				len(receivedLines), expectedLines, receivedLines)
		}
	}

	// Kill the process since we're done testing
	if err := cmd.Process.Kill(); err != nil {
		t.Errorf("Failed to kill process: %v", err)
	}
	cmd.Wait()
	// Wait for the output reader goroutine to finish
	wg.Wait()

	// Verify output matches expected
	assert.Equal(t, expectedLines, len(receivedLines))
	for i, line := range receivedLines {
		assert.Equal(t, fmt.Sprintf("test line %d", i+1), line)
	}
}
