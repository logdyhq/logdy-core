package utils

import "testing"

func TestStripAnsi(t *testing.T) {

	txt := []struct {
		in  string
		out string
	}{
		{
			in:  "This is a string with \033[31mred\033[0m text and \033[1mbold\033[0m formatting.",
			out: "This is a string with red text and bold formatting.",
		},
		{
			in:  "\033[22m\033[1m\033[36mApr 22 13:36:35\033[0m\033[22m\033[36m \033[0m\033[22m\033[1m\033[34mqra-team\033[0m\033[22m\033[36m",
			out: "Apr 22 13:36:35 qra-team",
		},
	}

	for _, line := range txt {
		stripped := StripAnsi(line.in)

		if stripped != line.out {
			t.Error(line.in, "doesnt equal", line.out)
		}
	}
}
