package modes

import (
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"testing"

	"github.com/logdyhq/logdy-core/utils"
	"github.com/stretchr/testify/assert" // Replace with your favorite testing framework
)

func TestUtilsCutByString(t *testing.T) {

	// Create a temporary file with some test data
	data := `This is line 1
This is line 2 (cut here)
This is line 3 to be copied
This is line 4 (end cut here)
This is line 5`
	tmpFile, err := ioutil.TempFile("", "cut_by_string_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up the temporary file

	_, err = tmpFile.Write([]byte(data))
	if err != nil {
		t.Fatal(err)
	}

	// Define test cases
	tests := []struct {
		name            string
		file            string
		start           string
		end             string
		caseInsensitive bool
		outFile         string
		expectedOutput  string
	}{
		{
			name:            "Basic cut (case sensitive)",
			file:            tmpFile.Name(),
			start:           "(cut here)",
			end:             "end cut",
			caseInsensitive: false,
			outFile:         "",
			expectedOutput:  "This is line 2 (cut here)\nThis is line 3 to be copied\nThis is line 4 (end cut here)\n",
		},
		{
			name:            "Basic cut (case sensitive)",
			file:            tmpFile.Name(),
			start:           "(cut HERE)",
			end:             "END cut",
			caseInsensitive: true,
			outFile:         "",
			expectedOutput:  "This is line 2 (cut here)\nThis is line 3 to be copied\nThis is line 4 (end cut here)\n",
		},
		{
			name:            "Basic cut (case sensitive)",
			file:            tmpFile.Name(),
			start:           "(cut here)",
			end:             "end cut",
			caseInsensitive: false,
			outFile:         t.TempDir() + "/cut_output.txt",
			expectedOutput:  "This is line 2 (cut here)\nThis is line 3 to be copied\nThis is line 4 (end cut here)\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var oldWriter = os.Stdout
			stdout, err := os.CreateTemp("", "")
			if err != nil {
				panic(err)
			}
			if tc.outFile == "" {
				os.Stdout = stdout
				defer func() {
					os.Stdout = oldWriter
				}()
			}

			UtilsCutByString(tc.file, tc.start, tc.end, tc.caseInsensitive, tc.outFile, "", 0)

			// Assertions
			if tc.outFile == "" {
				os.Stdout.Sync()
				fl, _ := os.Open(stdout.Name())
				content, err := io.ReadAll(fl)

				if err != nil {
					panic(err)
				}
				assert.Equal(t, tc.expectedOutput, string(content))
			} else {
				// Read output file content
				outputFileData, err := ioutil.ReadFile(tc.outFile)
				assert.Nil(t, err)
				assert.Equal(t, tc.expectedOutput, string(outputFileData))
			}
		})
	}
}

func TestUtilsCutByStringLong(t *testing.T) {

	// Create a temporary file with some test data
	const loremipsum = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua"

	tmpFile, err := ioutil.TempFile("", "cut_by_string_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up the temporary file

	for i := 0; i <= 10_000; i++ {
		_, err := tmpFile.WriteString(strconv.Itoa(i) + "_" + loremipsum + "\n")
		if err != nil {
			t.Fatal(err)
		}
	}

	// Define test cases
	tests := []struct {
		name            string
		file            string
		start           string
		end             string
		caseInsensitive bool
		outFile         string
		expectedLines   int
	}{
		{
			name:            "Basic cut (case sensitive)",
			file:            tmpFile.Name(),
			start:           "7650",
			end:             "9220",
			caseInsensitive: false,
			outFile:         "",
			expectedLines:   1571,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var oldWriter = os.Stdout
			stdout, err := os.CreateTemp("", "")
			if err != nil {
				panic(err)
			}
			if tc.outFile == "" {
				os.Stdout = stdout
				defer func() {
					os.Stdout = oldWriter
				}()
			}

			UtilsCutByString(tc.file, tc.start, tc.end, tc.caseInsensitive, tc.outFile, "", 0)

			// Assertions
			if tc.outFile == "" {
				os.Stdout.Sync()
				fl, _ := os.Open(stdout.Name())
				num, err := utils.LineCounter(fl)
				assert.Nil(t, err)

				if err != nil {
					panic(err)
				}
				assert.Equal(t, tc.expectedLines, num)
			} else {
				// Read output file content
				outputFileData, err := os.Open(tc.outFile)
				assert.Nil(t, err)
				num, err := utils.LineCounter(outputFileData)
				assert.Nil(t, err)
				assert.Equal(t, tc.expectedLines, num)
			}
		})
	}
}

func TestUtilsCutByStringTime(t *testing.T) {

	// Create a temporary file with some test data
	data := `[22:12:09] This is line 1
[22:12:10] This is line 2 (cut here)
[22:12:11] This is line 3 to be copied
[22:12:12] This is line 4 (end cut here)
[22:12:13] This is line 5`
	tmpFile, err := ioutil.TempFile("", "cut_by_string_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up the temporary file

	_, err = tmpFile.Write([]byte(data))
	if err != nil {
		t.Fatal(err)
	}

	// Define test cases
	tests := []struct {
		name            string
		file            string
		start           string
		end             string
		caseInsensitive bool
		outFile         string
		expectedOutput  string
	}{
		{
			name:            "Basic cut (case sensitive)",
			file:            tmpFile.Name(),
			start:           "22:12:10",
			end:             "22:12:12",
			caseInsensitive: false,
			outFile:         "",
			expectedOutput:  "[22:12:10] This is line 2 (cut here)\n[22:12:11] This is line 3 to be copied\n[22:12:12] This is line 4 (end cut here)\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var oldWriter = os.Stdout
			stdout, err := os.CreateTemp("", "")
			if err != nil {
				panic(err)
			}
			if tc.outFile == "" {
				os.Stdout = stdout
				defer func() {
					os.Stdout = oldWriter
				}()
			}

			UtilsCutByString(tc.file, tc.start, tc.end, tc.caseInsensitive, tc.outFile, "15:04:05", 1)

			// Assertions
			if tc.outFile == "" {
				os.Stdout.Sync()
				fl, _ := os.Open(stdout.Name())
				content, err := io.ReadAll(fl)

				if err != nil {
					panic(err)
				}
				assert.Equal(t, tc.expectedOutput, string(content))
			} else {
				// Read output file content
				outputFileData, err := ioutil.ReadFile(tc.outFile)
				assert.Nil(t, err)
				assert.Equal(t, tc.expectedOutput, string(outputFileData))
			}
		})
	}
}

func TestUtilsCutByStringDate(t *testing.T) {

	// Create a temporary file with some test data
	data := `[01/05/2024 22:12:09.100] This is line 1
[01/05/2024 22:12:10.110] This is line 2 (cut here)
[01/05/2024 22:12:10.120] This is line 3 to be copied
[01/05/2024 22:12:12.130] This is line 4 (end cut here)
[01/05/2024 22:12:12.140] This is line 5`
	tmpFile, err := ioutil.TempFile("", "cut_by_string_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up the temporary file

	_, err = tmpFile.Write([]byte(data))
	if err != nil {
		t.Fatal(err)
	}

	// Define test cases
	tests := []struct {
		name            string
		file            string
		start           string
		end             string
		caseInsensitive bool
		outFile         string
		expectedOutput  string
	}{
		{
			name:            "Basic cut (case sensitive)",
			file:            tmpFile.Name(),
			start:           "01/05/2024 22:12:10.000",
			end:             "01/05/2024 22:12:12.000",
			caseInsensitive: false,
			outFile:         "",
			expectedOutput: `[01/05/2024 22:12:10.110] This is line 2 (cut here)
[01/05/2024 22:12:10.120] This is line 3 to be copied
[01/05/2024 22:12:12.130] This is line 4 (end cut here)
`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var oldWriter = os.Stdout
			stdout, err := os.CreateTemp("", "")
			if err != nil {
				panic(err)
			}
			if tc.outFile == "" {
				os.Stdout = stdout
				defer func() {
					os.Stdout = oldWriter
				}()
			}

			UtilsCutByString(tc.file, tc.start, tc.end, tc.caseInsensitive, tc.outFile, "01/02/2006 15:04:05.000", 1)

			// Assertions
			if tc.outFile == "" {
				os.Stdout.Sync()
				fl, _ := os.Open(stdout.Name())
				content, err := io.ReadAll(fl)

				if err != nil {
					panic(err)
				}
				assert.Equal(t, tc.expectedOutput, string(content))
			} else {
				// Read output file content
				outputFileData, err := ioutil.ReadFile(tc.outFile)
				assert.Nil(t, err)
				assert.Equal(t, tc.expectedOutput, string(outputFileData))
			}
		})
	}
}

func TestUtilsCutByLineNumber(t *testing.T) {

	// Create a temporary file with some test data
	data := `This is line 1
This is line 2 (cut here)
This is line 3 to be copied
This is line 4 (end cut here)
This is line 5`
	tmpFile, err := ioutil.TempFile("", "cut_by_string_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up the temporary file

	_, err = tmpFile.Write([]byte(data))
	if err != nil {
		t.Fatal(err)
	}

	// Define test cases
	tests := []struct {
		name           string
		file           string
		count          int
		offset         int
		outFile        string
		expectedOutput string
	}{
		{
			name:           "Basic cut",
			file:           tmpFile.Name(),
			count:          2,
			offset:         2,
			outFile:        "",
			expectedOutput: "This is line 2 (cut here)\nThis is line 3 to be copied\n",
		},
		{
			name:           "Basic cut (to file)",
			file:           tmpFile.Name(),
			count:          2,
			offset:         3,
			outFile:        t.TempDir() + "/cut_output.txt",
			expectedOutput: "This is line 3 to be copied\nThis is line 4 (end cut here)\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var oldWriter = os.Stdout
			stdout, err := os.CreateTemp("", "")
			if err != nil {
				panic(err)
			}
			if tc.outFile == "" {
				os.Stdout = stdout
				defer func() {
					os.Stdout = oldWriter
				}()
			}

			UtilsCutByLineNumber(tc.file, tc.count, tc.offset, tc.outFile)

			// Assertions
			if tc.outFile == "" {
				os.Stdout.Sync()
				fl, _ := os.Open(stdout.Name())
				content, err := io.ReadAll(fl)

				if err != nil {
					panic(err)
				}
				assert.Equal(t, tc.expectedOutput, string(content))
			} else {
				// Read output file content
				outputFileData, err := ioutil.ReadFile(tc.outFile)
				assert.Nil(t, err)
				assert.Equal(t, tc.expectedOutput, string(outputFileData))
			}
		})
	}
}

func TestUtilsCutLineNumberLong(t *testing.T) {

	// Create a temporary file with some test data
	const loremipsum = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua"

	tmpFile, err := ioutil.TempFile("", "cut_by_string_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up the temporary file

	for i := 0; i <= 10_000; i++ {
		_, err := tmpFile.WriteString(strconv.Itoa(i) + "_" + loremipsum + "\n")
		if err != nil {
			t.Fatal(err)
		}
	}

	// Define test cases
	tests := []struct {
		name            string
		file            string
		count           int
		offset          int
		caseInsensitive bool
		outFile         string
	}{
		{
			name:            "Basic cut",
			file:            tmpFile.Name(),
			count:           3,
			offset:          7650,
			caseInsensitive: false,
			outFile:         "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var oldWriter = os.Stdout
			stdout, err := os.CreateTemp("", "")
			if err != nil {
				panic(err)
			}
			if tc.outFile == "" {
				os.Stdout = stdout
				defer func() {
					os.Stdout = oldWriter
				}()
			}

			UtilsCutByLineNumber(tc.file, tc.count, tc.offset, tc.outFile)

			// Assertions
			if tc.outFile == "" {
				os.Stdout.Sync()
				fl, _ := os.Open(stdout.Name())
				num, err := utils.LineCounter(fl)
				assert.Nil(t, err)

				if err != nil {
					panic(err)
				}
				assert.Equal(t, tc.count, num)
			} else {
				// Read output file content
				outputFileData, err := os.Open(tc.outFile)
				assert.Nil(t, err)
				num, err := utils.LineCounter(outputFileData)
				assert.Nil(t, err)
				assert.Equal(t, tc.count, num)
			}
		})
	}
}
