package utils

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	"github.com/logdyhq/logdy-core/models"
)

func TestProcessIncomingMessages_NoFileAppend(t *testing.T) {
	// Create a mock channel for incoming messages
	in := make(chan models.Message, 1)
	out := ProcessIncomingMessages(in, "", false)

	// Send a test message
	testMsg := models.Message{Content: "Hello World"}
	in <- testMsg

	// Read the message from the output channel
	receivedMsg := <-out

	// Assert that the received message is the same as the sent message
	if receivedMsg.Content != testMsg.Content {
		t.Errorf("Expected message %v, got %v", testMsg, receivedMsg)
	}
}

func TestProcessIncomingMessages_AppendFileJSON(t *testing.T) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "process_messages_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Process messages with file append (JSON)
	in := make(chan models.Message, 1)
	out := ProcessIncomingMessages(in, tmpfile.Name(), false)

	// Send a test message
	testMsg := models.Message{Content: "Test message"}
	in <- testMsg

	log.Println(tmpfile.Name())

	time.Sleep(10 * time.Millisecond) //need to wait for the write to be flushed to a file
	// Read the content from the temporary file
	data, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Expected JSON encoded message
	expected, err := json.Marshal(testMsg)
	if err != nil {
		t.Fatal(err)
	}
	expected = append(expected, '\n')

	// Assert that the written content matches the expected JSON
	if !bytes.Equal(data, expected) {
		t.Errorf("Expected file content %v, got %v", string(expected), string(data))
	}

	// Assert that the message is passed through the channel
	receivedMsg := <-out
	if receivedMsg.Content != testMsg.Content {
		t.Errorf("Expected message %v, got %v", testMsg, receivedMsg)
	}
}

func TestProcessIncomingMessages_AppendFileRaw(t *testing.T) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "process_messages_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Process messages with file append (JSON)
	in := make(chan models.Message, 1)
	out := ProcessIncomingMessages(in, tmpfile.Name(), true)

	// Send a test message
	testMsg := models.Message{Content: "Test message"}
	in <- testMsg

	log.Println(tmpfile.Name())

	time.Sleep(10 * time.Millisecond) //need to wait for the write to be flushed to a file
	// Read the content from the temporary file
	data, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Expected JSON encoded message
	expected := testMsg.Content + "\n"

	// Assert that the written content matches the expected JSON
	if !bytes.Equal(data, []byte(expected)) {
		t.Errorf("Expected file content %v, got %v", string(expected), string(data))
	}

	// Assert that the message is passed through the channel
	receivedMsg := <-out
	if receivedMsg.Content != testMsg.Content {
		t.Errorf("Expected message %v, got %v", testMsg, receivedMsg)
	}
}
