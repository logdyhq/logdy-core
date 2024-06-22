package modes

import (
	"context"
	"encoding/json"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/logdyhq/logdy-core/models"
	"github.com/logdyhq/logdy-core/utils"
)

var correlationIds []string

func init() {
	generateCorrelationIds()
}

func generateCorrelationIds() {
	correlationIds = []string{""}
	for i := 0; i <= 4; i++ {
		correlationIds = append(correlationIds, utils.RandStringRunes(8))
	}
}

func GenerateRandomData(jsonFormat bool, numPerSec int, ch chan models.Message, ctx context.Context) {

	if numPerSec > 100 {
		numPerSec = 100
	}

	if numPerSec <= 0 {
		return // produce no data, so just leave
	}

	i := 0
	for {
		i++

		if i%(60*numPerSec) == 0 {
			generateCorrelationIds()
		}

		if ctx.Err() != nil {
			return
		}

		var msg string
		if jsonFormat {
			msg = generateJsonRandomData()
		} else {
			msg = generateTextRandomData()
		}

		mo := models.MessageOrigin{}

		if rand.Intn(100) >= 50 {
			mo.File = utils.PickRandom[string]([]string{"foo1.log", "foo2.log", "foo3.log"})
		} else {
			mo.Port = utils.PickRandom[string]([]string{"4356", "4333", "4262"})
		}
		if rand.Intn(100) >= 90 {
			mo.File = ""
			mo.Port = ""
		}

		produce(ch, msg, models.MessageTypeStdout, &mo)
		time.Sleep(time.Duration((1 / float64(numPerSec)) * float64(time.Second)))
	}

}

func generateTextRandomData() string {
	return strings.Join([]string{
		time.Now().Format("15:04:05.0000"),
		strconv.Itoa(int(time.Now().UnixMilli())),
		strconv.Itoa(rand.Intn(500-300) + 300),
		gofakeit.UUID(),
		gofakeit.DomainName(),
		gofakeit.IPv4Address(),
		gofakeit.URL(),
		gofakeit.LogLevel("log"),
		gofakeit.UserAgent(),
		gofakeit.HTTPMethod(),
		utils.PickRandom[string](correlationIds),
	}, " | ")
}

func generateJsonRandomData() string {
	val, _ := json.Marshal(map[string]string{
		"ts":             time.Now().Format("15:04:05.0000"),
		"unix":           strconv.Itoa(int(time.Now().UnixMilli())),
		"duration":       strconv.Itoa(rand.Intn(500-300) + 300),
		"uuid":           gofakeit.UUID(),
		"domain":         gofakeit.DomainName(),
		"ipv4":           gofakeit.IPv4Address(),
		"url":            gofakeit.URL(),
		"level":          gofakeit.LogLevel("log"),
		"ua":             gofakeit.UserAgent(),
		"method":         gofakeit.HTTPMethod(),
		"correlation_id": utils.PickRandom[string](correlationIds),
	})

	return string(val)
}
