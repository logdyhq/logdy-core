package utils

import (
	"bytes"
	"context"
	"io"
	"os"

	"github.com/cheggaaa/pb/v3"
)

type Line struct {
	seq  int
	Line []byte
}

func LineCounterWithChannel(r io.Reader, fn func(line Line, cancel func())) error {
	ctx, cancel := context.WithCancel(context.Background())
	const bufferSize = 64 * 1024
	buf := make([]byte, bufferSize)
	var previousLine = []byte{}
	seq := 0

	for {

		if ctx.Err() != nil {
			cancel()
			return nil
		}

		i, err := r.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}

		if i == 0 && len(previousLine) == 0 {
			cancel()
			return nil
		}

		previousLine = append(previousLine, buf[:i]...)
		newlineIndex := bytes.IndexByte(previousLine, '\n')

		if newlineIndex == -1 {

			if i == bufferSize {
				continue
			}

			if err == io.EOF || (i == 0 || i < bufferSize) {
				fn(Line{seq: seq, Line: previousLine}, cancel)
				return nil
			}
		}

		lines := bytes.Split(previousLine, []byte{'\n'})
		previousLine = lines[len(lines)-1]
		lines = lines[:len(lines)-1]

		for _, line := range lines {
			seq++
			fn(Line{seq: seq, Line: line}, cancel)
		}

		if err == io.EOF {
			fn(Line{seq: seq, Line: previousLine}, cancel)
			return nil
		}
	}
}

func OpenFileForReading(file string) (io.Reader, int64, *pb.ProgressBar) {
	reader, err := os.Open(file)

	if err != nil {
		panic(err)
	}

	fi, _ := reader.Stat()
	bar := pb.Full.Start64(fi.Size())
	return bar.NewProxyReader(reader), fi.Size(), bar
	// return reader, fi.Size()
}
