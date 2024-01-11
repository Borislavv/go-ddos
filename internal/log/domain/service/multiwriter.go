package logservice

import "io"

type MultiWriter struct {
	writers []io.Writer
}

func NewMultiWriter(writers ...io.Writer) *MultiWriter {
	return &MultiWriter{
		writers: writers,
	}
}

func (w *MultiWriter) Write(p []byte) (n int, err error) {
	for _, writer := range w.writers {
		l, e := writer.Write(p)
		if n == 0 {
			n = l
		}
		if e != nil {
			err = e
		}
	}
	return n, err
}
