package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/logdyhq/logdy-core/models"
)

func TestUnmarshalJSONTimestamp(t *testing.T) {
	tests := []struct {
		input          string
		expectedTime   time.Time
		expectingError bool
	}{
		{
			input:        `"2023-07-08T14:00:00Z"`,
			expectedTime: time.Date(2023, 7, 8, 14, 0, 0, 0, time.UTC),
		},
		{
			input:        `"1688815200000"`,
			expectedTime: time.Unix(0, 1688815200000*int64(time.Millisecond)),
		},
		{
			input:          `"invalid"`,
			expectingError: true,
		},
		{
			input:          ``,
			expectingError: false,
		},
	}

	for _, test := range tests {
		var ts Timestamp
		err := ts.UnmarshalJSON([]byte(test.input))
		if test.expectingError {
			if err == nil {
				t.Errorf("expected error for input %s, but got nil", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error for input %s: %v", test.input, err)
			} else {
				if test.input == "" && ts.Time.IsZero() {
					t.Errorf("for empty, expected current time, but got empty time")
				}
				if test.input != "" && !ts.Time.Equal(test.expectedTime) {
					t.Errorf("for input %s, expected %v, but got %v", test.input, test.expectedTime, ts.Time)
				}
			}
		}
	}
}

func TestUnmarshalJSONLog(t *testing.T) {
	tests := []struct {
		input          string
		expected       string
		expectingError bool
	}{
		{
			input:    `blah`,
			expected: `blah`,
		},
		{
			input:    `{"asd": "foo"}`,
			expected: `{"asd": "foo"}`,
		},
		{
			input:    ``,
			expected: ``,
		},
	}

	for _, test := range tests {
		var msg LogMessage
		err := msg.UnmarshalJSON([]byte(test.input))
		if test.expectingError {
			if err == nil {
				t.Errorf("expected error for input %s, but got nil", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error for input %s: %v", test.input, err)
			} else if msg.String != test.expected {
				t.Errorf("for input %s, expected %v, but got %v", test.input, test.expected, msg.String)
			}
		}
	}
}

func TestUnmarshalLogRequest(t *testing.T) {
	jsonStr := `{
		"logs": [
			{"ts": "2023-07-08T14:00:00Z"},
			{"ts": "1688815200000"}
		]
	}`

	var logRequest LogRequest
	err := json.Unmarshal([]byte(jsonStr), &logRequest)
	if err != nil {
		t.Fatalf("unexpected error unmarshalling json: %v", err)
	}

	if len(logRequest.Logs) != 2 {
		t.Fatalf("expected 2 log entries, got %d", len(logRequest.Logs))
	}

	expectedTimes := []time.Time{
		time.Date(2023, 7, 8, 14, 0, 0, 0, time.UTC),
		time.Unix(0, 1688815200000*int64(time.Millisecond)),
	}

	for i, log := range logRequest.Logs {
		if !log.Ts.Time.Equal(expectedTimes[i]) {
			t.Errorf("log %d: expected %v, but got %v", i, expectedTimes[i], log.Ts.Time)
		}
	}
}

func TestHandleLog(t *testing.T) {
	tests := []struct {
		name           string
		input          LogRequest
		expectedStatus int
		expectedMsgs   int
	}{
		{
			name: "Valid request with single log",
			input: LogRequest{
				Logs: []LogItemRequest{
					{
						Ts:  Timestamp{Time: time.Now()},
						Log: LogMessage{String: "Test log message"},
					},
				},
				Source: "test-client",
			},
			expectedStatus: http.StatusAccepted,
			expectedMsgs:   1,
		},
		{
			name: "Valid request with multiple logs",
			input: LogRequest{
				Logs: []LogItemRequest{
					{
						Ts:  Timestamp{Time: time.Now()},
						Log: LogMessage{String: "Log message 1"},
					},
					{
						Ts:  Timestamp{Time: time.Now()},
						Log: LogMessage{String: "Log message 2"},
					},
				},
				Source: "test-client",
			},
			expectedStatus: http.StatusAccepted,
			expectedMsgs:   2,
		},
		{
			name:           "Invalid request (empty body)",
			input:          LogRequest{},
			expectedStatus: http.StatusAccepted,
			expectedMsgs:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a channel to receive messages
			msgChan := make(chan models.Message, 10)

			// Create a request with the test input
			body, _ := json.Marshal(tt.input)
			req, _ := http.NewRequest("POST", "/log", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Call the handler function
			handler := handleLog(msgChan)
			handler(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// Check the number of messages sent to the channel
			receivedMsgs := 0
			for i := 0; i < tt.expectedMsgs; i++ {
				select {
				case msg := <-msgChan:
					receivedMsgs++
					if msg.Mtype != models.MessageTypeStdout {
						t.Errorf("unexpected message type: got %v, want %v", msg.MessageType, models.MessageTypeStdout)
					}
					if msg.Origin.ApiSource != tt.input.Source {
						t.Errorf("unexpected API client ID: got %v, want %v", msg.Origin.ApiSource, tt.input.Source)
					}
				default:
					t.Errorf("expected %d messages, got %d", tt.expectedMsgs, receivedMsgs)
					return
				}
			}

			// Ensure no extra messages were sent
			select {
			case <-msgChan:
				t.Errorf("received unexpected extra message")
			default:
				// This is the expected case
			}
		})
	}
}
