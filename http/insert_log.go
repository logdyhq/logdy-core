package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/logdyhq/logdy-core/models"
	"github.com/logdyhq/logdy-core/modes"
	"github.com/logdyhq/logdy-core/utils"
)

type LogItemRequest struct {
	Ts  Timestamp  `json:"ts"`
	Log LogMessage `json:"log"`
}

type LogRequest struct {
	Logs   []LogItemRequest `json:"logs"`
	Source string           `json:"source"`
}

func handleLog(messageChannel chan models.Message) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			httpError("This endpoint accepts only POST method", w, http.StatusMethodNotAllowed)
			return
		}

		var p LogRequest
		err := json.NewDecoder(r.Body).Decode(&p)

		if err != nil {
			httpError(err.Error(), w, http.StatusInternalServerError)
			return
		}

		utils.Logger.Debugf("Inserting a batch of log messages (%d)", len(p.Logs))
		for _, el := range p.Logs {
			modes.ProduceMessageStringTimestamped(messageChannel, el.Log.String, models.MessageTypeStdout, &models.MessageOrigin{
				ApiSource: p.Source,
			}, el.Ts.Time)
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

type LogMessage struct {
	String string
}

func (t *LogMessage) UnmarshalJSON(data []byte) error {
	t.String = string(data)
	return nil
}

// Define a custom type for handling the timestamp
type Timestamp struct {
	time.Time
}

func (t *Timestamp) UnmarshalJSON(data []byte) error {
	// Remove the quotes from the JSON string
	s := string(data)
	if s == "null" {
		return nil
	}

	if s == "" {
		t.Time = time.Now()
		return nil
	}

	s = s[1 : len(s)-1]

	// Try to parse as RFC 3339
	if parsedTime, err := time.Parse(time.RFC3339, s); err == nil {
		t.Time = parsedTime
		return nil
	}

	// Try to parse as UNIX timestamp in milliseconds
	if ms, err := strconv.ParseInt(s, 10, 64); err == nil {
		t.Time = time.Unix(0, ms*int64(time.Millisecond))
		return nil
	}

	return fmt.Errorf("invalid timestamp format")
}
