package modes

import (
	"io"
	"logdy/utils"
	"os"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/sirupsen/logrus"
)

func UtilsCutByString(file string, start string, end string, caseInsensitive bool, outFile string, dateFormat string, searchOffset int) {
	_, err := os.Stat(file)
	if err != nil {
		utils.Logger.WithFields(logrus.Fields{
			"path":  file,
			"error": err.Error(),
		}).Error("Reading file failed")
		return
	}

	var size int64
	var bar *pb.ProgressBar
	var r io.Reader
	if outFile == "" {
		r, size = utils.OpenFileForReading(file)
	} else {
		r, size, bar = utils.OpenFileForReadingWithProgress(file)
	}

	utils.Logger.WithFields(logrus.Fields{
		"path":       file,
		"size_bytes": size,
	}).Info("Reading file")

	if caseInsensitive {
		start = strings.ToLower(start)
		end = strings.ToLower(end)
	}

	var f *os.File
	if outFile != "" {
		f, err = os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer f.Close() // Close the file when we're done
	}

	var startDate time.Time
	var endDate time.Time

	if dateFormat != "" {
		startDate, err = time.Parse(dateFormat, start)
		if err != nil {
			panic(err)
		}
		endDate, err = time.Parse(dateFormat, end)
		if err != nil {
			panic(err)
		}
	}

	started := false
	stopped := false
	utils.LineCounterWithChannel(r, func(line utils.Line, cancel func()) {
		if stopped {
			return
		}
		ln := string(line.Line)

		if caseInsensitive {
			ln = strings.ToLower(ln)
		}

		if dateFormat == "" && strings.Contains(ln, start) {
			started = true
		}

		if dateFormat != "" {
			t, err := time.Parse(dateFormat, ln[searchOffset:searchOffset+len(dateFormat)])

			if err != nil {
				panic(err)
			}

			if !t.IsZero() && (t.After(startDate) || t.Equal(startDate)) {
				started = true
			}
		}

		if !started {
			return
		}

		if outFile == "" {
			os.Stdout.Write(line.Line)
			os.Stdout.Write([]byte{'\n'})
		} else {
			f.Write(line.Line)
			f.Write([]byte{'\n'})
			if err != nil {
				panic(err)
			}
		}

		if dateFormat == "" && strings.Contains(ln, end) {
			cancel()
			stopped = true
			return
		}

		if dateFormat != "" {
			t, err := time.Parse(dateFormat, ln[searchOffset:searchOffset+len(dateFormat)])

			if err != nil {
				panic(err)
			}

			if !t.IsZero() && (t.After(endDate) || t.Equal(endDate)) {
				cancel()
				stopped = true
				return
			}
		}
	})

	if outFile != "" {
		bar.Finish()
	}
}
