package utils

import (
	"bytes"
	"fmt"
	"io"
)

// PrefixWriter prefixes each line with a tag
type PrefixWriter struct {
	Prefix string
	Writer io.Writer
}

func (w *PrefixWriter) Write(p []byte) (n int, err error) {
	lines := bytes.Split(p, []byte("\n"))
	for i, line := range lines {
		if len(line) == 0 && i == len(lines)-1 {
			continue
		}
		out := fmt.Sprintf("[%s] %s\n", w.Prefix, string(line))
		w.Writer.Write([]byte(out))
	}
	return len(p), nil
}
