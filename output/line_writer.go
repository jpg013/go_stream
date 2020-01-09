package output

import (
	"fmt"
	"github.com/jpg013/go_stream/types"
	"io"
	"sync/atomic"
)

// LineWriter is an output type that writes messages to an io.WriterCloser type as lines.
type LineWriter struct {
	running     int32
	customDelim []byte
	handle      io.WriteCloser
	inputChan   chan types.Chunk
	closeOnExit bool
	closeChan   chan struct{}
}

// NewLineWriter creates a new LineWriter output type.
func NewLineWriter(
	handle io.WriteCloser,
	closeOnExit bool,
	customDelimiter []byte,
) *LineWriter {
	return &LineWriter{
		running:     1,
		customDelim: customDelimiter,
		closeOnExit: closeOnExit,
		handle:      handle,
		closeChan:   make(chan struct{}),
	}
}

// Closed returns bool indicating whether line writer is closed or writing
func (w *LineWriter) Closed() bool {
	return atomic.LoadInt32(&w.running) != 1
}

// Consume assigns an input channel for the output to read and starts the loop
func (w *LineWriter) Consume(ch chan types.Chunk) error {
	if w.inputChan != nil {
		// return types.ErrAlreadyStarted
		return types.ErrAlreadyStarted
	}
	w.inputChan = ch
	go w.loop()
	return nil
}

// loop is an internal loop that brokers incoming messages to output pipe.
func (w *LineWriter) loop() {
	defer func() {
		if w.closeOnExit {
			w.handle.Close()
		}
	}()

	delim := []byte("\n")
	if len(w.customDelim) > 0 {
		delim = w.customDelim
	}

	for atomic.LoadInt32(&w.running) == 1 {
		var chunk types.Chunk
		var more bool
		select {
		case chunk, more = <-w.inputChan:
			if !more {
				return
			}
		case <-w.closeChan:
			return
		}

		_, err := fmt.Fprintf(w.handle, "%s%s", chunk, delim)

		if err != nil {
			panic(err)
		}
	}
}

// CloseAsync shuts down the File output and stops processing messages.
func (w *LineWriter) CloseAsync() {
	if atomic.CompareAndSwapInt32(&w.running, 1, 0) {
		close(w.closeChan)
	}
}
