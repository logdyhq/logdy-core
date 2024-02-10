package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
)

var FallthroughGlobal = false

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
		if mo.Port != "" {
			fields["origin_port"] = mo.Port
		}
		if mo.File != "" {
			fields["origin_file"] = mo.File
		}
	}

	logger.WithFields(fields).Debug("Producing message")

	if FallthroughGlobal {
		fmt.Println(line)
	}

	ch <- Message{
		Mtype:       mt,
		Content:     line,
		JsonContent: cs,
		IsJson:      validJson == nil,
		BaseMessage: BaseMessage{MessageType: "log"},
		Origin:      mo,
		Ts:          time.Now(),
	}
}

func trunc(str string, limit int) string {

	if len(str) <= limit {
		return str
	}

	return str[:limit] + "..."
}
