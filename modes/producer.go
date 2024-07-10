package modes

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/logdyhq/logdy-core/models"
	"github.com/logdyhq/logdy-core/utils"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
)

var FallthroughGlobal = false
var DisableANSICodeStripping = false

func ProduceMessageStringTimestamped(ch chan models.Message, line string, mt models.LogType, mo *models.MessageOrigin, ts time.Time) {

	if !DisableANSICodeStripping {
		line = utils.StripAnsi(line)
	}

	validJson := fastjson.Validate(line)
	var cs json.RawMessage
	if validJson == nil {
		cs = json.RawMessage(line)
	}

	fields := logrus.Fields{
		"line": utils.Trunc(line, 45),
	}
	if mo != nil {
		if mo.Port != "" {
			fields["origin_port"] = mo.Port
		}
		if mo.File != "" {
			fields["origin_file"] = mo.File
		}
	}

	utils.Logger.WithFields(fields).Debug("Producing message")

	if FallthroughGlobal {
		if mt == models.MessageTypeStdout {
			fmt.Fprintln(os.Stdout, line)
		}
		if mt == models.MessageTypeStderr {
			fmt.Fprintln(os.Stderr, line)
		}
	}

	ch <- models.Message{
		Id:          strconv.FormatInt(time.Now().UnixMicro(), 10),
		Mtype:       mt,
		Content:     line,
		JsonContent: cs,
		IsJson:      validJson == nil,
		BaseMessage: models.BaseMessage{MessageType: "log"},
		Origin:      mo,
		Ts:          ts.UnixMilli(),
	}
}

func ProduceMessageString(ch chan models.Message, line string, mt models.LogType, mo *models.MessageOrigin) {
	ProduceMessageStringTimestamped(ch, line, mt, mo, time.Now())
}
