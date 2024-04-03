package utils

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLineCounterWithChannel(t *testing.T) {

	tests := []struct {
		input      string
		buffer     int
		linesCount int
	}{
		{input: "123", buffer: 5, linesCount: 1},
		{input: "123456", buffer: 5, linesCount: 1},
		{input: "123456790abc", buffer: 5, linesCount: 1},

		{input: "12\n3", buffer: 5, linesCount: 2},
		{input: "12\n3456789", buffer: 5, linesCount: 2},
		{input: "12\n3456\n789", buffer: 5, linesCount: 3},
		{input: "12\n34aaabbb56\n789", buffer: 5, linesCount: 3},
		{input: "12\n34aaabbbcccddd56\n789", buffer: 5, linesCount: 3},

		{input: "1\n\n2", buffer: 5, linesCount: 3},
		{input: "1\n\n\n\n2", buffer: 5, linesCount: 5},
	}

	for _, tc := range tests {

		chars := ""
		c := 0
		LineCounterWithChannel(bytes.NewBufferString(tc.input), func(line Line, cancel func()) {
			c++
			chars = chars + string(line.Line)
		})

		assert.Equal(t, tc.linesCount, c)
		assert.Equal(t, strings.ReplaceAll(tc.input, "\n", ""), chars)

	}

}
