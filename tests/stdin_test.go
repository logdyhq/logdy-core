package tests

import (
	"bufio"
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func runCmd(cmds []string, t *testing.T) {
	// Start logdy process in stdin mode with fallthrough enabled
	// -t enables fallthrough so we can see the output
	cmd := exec.Command("go", cmds...)

	// Get stdin pipe
	stdin, err := cmd.StdinPipe()
	assert.NoError(t, err)

	// Get stdout pipe
	stdout, err := cmd.StdoutPipe()
	assert.NoError(t, err)

	// Start the process
	err = cmd.Start()
	assert.NoError(t, err)

	// Create a channel to collect output lines
	outputChan := make(chan string)

	// Start goroutine to read stdout
	go func() {
		reader := bufio.NewReader(stdout)
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				close(outputChan)
				return
			}
			if err != nil {
				t.Errorf("Error reading stdout: %v", err)
				close(outputChan)
				return
			}
			// Only collect lines that contain our test data
			if strings.Contains(line, "test line") {
				outputChan <- strings.TrimSpace(line)
			}
		}
	}()

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

	// Collect output with timeout
	receivedLines := make([]string, 0)
	timeout := time.After(5 * time.Second)

	for i := 0; i < len(testLines); i++ {
		select {
		case line, ok := <-outputChan:
			if !ok {
				t.Fatal("Output channel closed before receiving all expected lines")
			}
			receivedLines = append(receivedLines, line)
		case <-timeout:
			t.Fatal("Timeout waiting for output")
		}
	}

	// Kill the process since we're done testing
	if err := cmd.Process.Kill(); err != nil {
		t.Errorf("Failed to kill process: %v", err)
	}

	// Verify output matches input
	assert.Equal(t, len(testLines), len(receivedLines))
	for i, testLine := range testLines {
		assert.Contains(t, receivedLines[i], testLine)
	}
}

func TestLogdyE2E_NoCommand(t *testing.T) {
	runCmd([]string{"run", "../main.go", "-t"}, t)
}

func TestLogdyE2E_StdinCommand(t *testing.T) {
	runCmd([]string{"run", "../main.go", "stdin", "-t"}, t)
}
