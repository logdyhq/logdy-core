package main

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
)

func produce(ch chan Message, line string, mt LogType, mo *MessageOrigin) {
	validJson := fastjson.Validate(line)
	var cs json.RawMessage
	if validJson == nil {
		cs = json.RawMessage(line)
	}

	fields := logrus.Fields{
		"line": trunc(line, 45),
	}
	if mo != nil {
		fields["origin_port"] = mo.Port
	}

	logger.WithFields(fields).Debug("Producing message")

	ch <- Message{
		Mtype:       mt,
		Content:     line,
		JsonContent: cs,
		IsJson:      validJson == nil,
		BaseMessage: BaseMessage{MessageType: "log"},
		Origin:      mo,
	}
}

func trunc(str string, limit int) string {

	if len(str) <= limit {
		return str
	}

	return str[:limit] + "..."
}
