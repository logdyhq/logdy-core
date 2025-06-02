package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/logdyhq/logdy-core/models"
)

func TestProcessIncomingMessagesWithRotation_BasicRotation(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "process_rotation_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "test.log")

	// Create channels
	in := make(chan models.Message, 10)

	// Small max size to trigger rotation easily
	maxSize := int64(50)
	maxBackups := 2

	// Start processing with rotation
	out := ProcessIncomingMessagesWithRotation(in, logFile, false, maxSize, maxBackups)

	// Send messages that will trigger rotation
	msg1 := models.Message{Content: "First message that is quite long"}
	msg2 := models.Message{Content: "Second message that will trigger rotation"}
	msg3 := models.Message{Content: "Third message after rotation"}

	in <- msg1
	time.Sleep(50 * time.Millisecond) // Allow time for writing

	in <- msg2
	time.Sleep(50 * time.Millisecond) // Allow time for rotation

	in <- msg3
	time.Sleep(50 * time.Millisecond) // Allow time for writing

	// Verify messages are passed through
	receivedCount := 0
	timeout := time.After(100 * time.Millisecond)

Loop:
	for {
		select {
		case <-out:
			receivedCount++
			if receivedCount == 3 {
				break Loop
			}
		case <-timeout:
			break Loop
		}
	}

	if receivedCount != 3 {
		t.Errorf("Expected to receive 3 messages, got %d", receivedCount)
	}

	// Check that rotation occurred
	backup1 := strings.Replace(logFile, ".log", ".1.log", 1)
	if _, err := os.Stat(backup1); os.IsNotExist(err) {
		t.Error("Expected backup file to exist after rotation")
	}

	// Verify current log file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Expected main log file to exist")
	}
}

func TestProcessIncomingMessagesWithRotation_RawMode(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "process_rotation_test_raw")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "test_raw.log")

	// Create channels
	in := make(chan models.Message, 5)

	// Use rotation with raw mode
	maxSize := int64(30)
	maxBackups := 1

	// Start processing with rotation in raw mode
	out := ProcessIncomingMessagesWithRotation(in, logFile, true, maxSize, maxBackups)

	// Send messages
	msg1 := models.Message{Content: "Raw message one"}
	msg2 := models.Message{Content: "Raw message two that triggers rotation"}

	in <- msg1
	time.Sleep(50 * time.Millisecond)

	in <- msg2
	time.Sleep(50 * time.Millisecond)

	// Read from output channel
	<-out
	<-out

	// Read current file content
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}

	// In raw mode, content should not be JSON
	var testJSON map[string]interface{}
	if err := json.Unmarshal(content, &testJSON); err == nil {
		t.Error("Expected raw content, but got valid JSON")
	}

	// Content should contain raw message
	if !strings.Contains(string(content), "Raw message") {
		t.Error("Expected raw message content in file")
	}
}

func TestProcessIncomingMessagesWithRotation_NoRotation(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "process_no_rotation_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "test_no_rotation.log")

	// Create channels
	in := make(chan models.Message, 5)

	// maxSize = 0 means no rotation
	out := ProcessIncomingMessagesWithRotation(in, logFile, false, 0, 0)

	// Send multiple messages
	for i := 0; i < 5; i++ {
		msg := models.Message{Content: "Message without rotation"}
		in <- msg
	}

	time.Sleep(100 * time.Millisecond)

	// Read messages from output
	for i := 0; i < 5; i++ {
		<-out
	}

	// Check that no backup files were created
	backup1 := strings.Replace(logFile, ".log", ".1.log", 1)
	if _, err := os.Stat(backup1); !os.IsNotExist(err) {
		t.Error("No backup files should exist when rotation is disabled")
	}

	// Verify all messages are in the main file
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 5 {
		t.Errorf("Expected 5 lines in log file, got %d", len(lines))
	}
}

func TestProcessIncomingMessages_BackwardCompatibility(t *testing.T) {
	// Test that the original function still works
	tmpDir, err := os.MkdirTemp("", "process_backward_compat_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	logFile := filepath.Join(tmpDir, "test_compat.log")

	// Create channels
	in := make(chan models.Message, 1)

	// Use original function (should use no rotation)
	out := ProcessIncomingMessages(in, logFile, false)

	// Send a message
	msg := models.Message{Content: "Backward compatible message"}
	in <- msg

	time.Sleep(50 * time.Millisecond)

	// Receive message
	received := <-out
	if received.Content != msg.Content {
		t.Errorf("Expected message content %q, got %q", msg.Content, received.Content)
	}

	// Verify file was created
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Error("Expected log file to be created")
	}
}
