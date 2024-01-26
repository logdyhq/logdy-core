package main

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
)

func produce(ch chan Message, line string, mt MessageType) {
	validJson := fastjson.Validate(line)
	var cs json.RawMessage
	if validJson == nil {
		cs = json.RawMessage(line)
	}

	logger.WithFields(logrus.Fields{
		"line": line,
	}).Debug("Producing message")

	ch <- Message{Mtype: mt, Content: line, JsonContent: cs, IsJson: validJson == nil}
}
