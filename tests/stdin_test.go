package tests

import (
	"bufio"
	"io"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func runCmd(cmds []string, t *testing.T) {
	// Start logdy process in stdin mode with fallthrough enabled
	// -t enables fallthrough so we can see the output
	cmd := exec.Command("go", cmds...)

	// Set process group for clean shutdown
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

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
	done := make(chan bool)

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
				// Ignore pipe errors since we're killing the process
				if !strings.Contains(err.Error(), "pipe") {
					t.Errorf("Error reading stdout: %v", err)
				}
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

	// Close stdin to signal we're done writing
	stdin.Close()

	// Collect output with timeout
	receivedLines := make([]string, 0)
	timeout := time.After(5 * time.Second)

	go func() {
		for i := 0; i < len(testLines); i++ {
			select {
			case line, ok := <-outputChan:
				if !ok {
					return
				}
				receivedLines = append(receivedLines, line)
				if len(receivedLines) == len(testLines) {
					done <- true
					return
				}
			case <-timeout:
				done <- true
				return
			}
		}
	}()

	<-done

	// Kill the process group to ensure clean shutdown
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err == nil {
		syscall.Kill(-pgid, syscall.SIGTERM)
	}

	// Wait with timeout to avoid hanging
	waitChan := make(chan error)
	go func() {
		waitChan <- cmd.Wait()
	}()

	select {
	case <-waitChan:
		// Process exited
	case <-time.After(2 * time.Second):
		// Force kill if it didn't exit cleanly
		cmd.Process.Kill()
		cmd.Wait()
	}

	// Verify output matches input
	assert.Equal(t, len(testLines), len(receivedLines))
	for i, testLine := range testLines {
		assert.Contains(t, receivedLines[i], testLine)
	}
}

func TestLogdyE2E_NoCommand(t *testing.T) {
	runCmd([]string{"run", "../.", "-t"}, t)
}

func TestLogdyE2E_StdinCommand(t *testing.T) {
	runCmd([]string{"run", "../.", "stdin", "-t"}, t)
}
