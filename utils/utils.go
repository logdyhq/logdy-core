package utils

import (
	"io"
	"math/rand"
	"os"
	"regexp"
	"time"
)

func LoadFile(configFilePath string) string {
	f, err := os.OpenFile(configFilePath, os.O_RDONLY, 0644)

	if err != nil {
		Logger.Error("Error while loading config file")
		panic(err)
	}

	bytes, err := io.ReadAll(f)

	if err != nil {
		Logger.Error("Error while loading config file")
		panic(err)
	}

	return string(bytes)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func Trunc(str string, limit int) string {

	if len(str) <= limit {
		return str
	}

	return str[:limit] + "..."
}

func PickRandom[T any](slice []T) T {
	if len(slice) == 0 {
		panic("slice is empty")
	}
	index := rand.Intn(len(slice))
	return slice[index]
}

func StripAnsi(str string) string {
	/**
	Regular expression to match ANSI escape sequences
	Regular Expression: We define a regular expression pattern using regexp.MustCompile. This pattern matches:
	\x1B: Escape character (ASCII code 27)
	\[: Opening square bracket
	[0-?]*: Zero or more occurrences of characters between 0 and ? (for parameter sequences)
	[ -/]*: Zero or more occurrences of spaces, hyphens, or forward slashes (for intermediate bytes)
	[@-~]: A single character between @ and ~ (for final byte)

	This approach uses a simplified regular expression that might not capture all possible ANSI escape sequences.
	For more comprehensive removal, you might need a more complex regular expression or consider
	using a dedicated library like "https://pkg.go.dev/github.com/pborman/ansi"
	*/
	ansiEscape := regexp.MustCompile(`\x1B\[[0-?]*[ -/]*[@-~]`)
	return ansiEscape.ReplaceAllString(str, "")
}
