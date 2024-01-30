package main

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
)

func produce(ch chan Message, line string, mt LogType) {
	validJson := fastjson.Validate(line)
	var cs json.RawMessage
	if validJson == nil {
		cs = json.RawMessage(line)
	}

	logger.WithFields(logrus.Fields{
		"line": trunc(line, 45),
	}).Debug("Producing message")

	ch <- Message{Mtype: mt, Content: line, JsonContent: cs, IsJson: validJson == nil, BaseMessage: BaseMessage{MessageType: "log"}}
}

func trunc(str string, limit int) string {
	if len(str) <= limit {
		return str
	}

	return str[:limit] + "..."
}
