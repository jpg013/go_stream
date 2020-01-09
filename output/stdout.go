package output

import (
	"os"
	"stream/types"
)

// STDOUTConfig contains configuration fields for the stdout based output type.
type STDOUTConfig struct {
	Delim string `json:"delimiter" yaml:"delimiter"`
}

// NewSTDOUTConfig creates a new STDOUTConfig with default values.
func NewSTDOUTConfig() STDOUTConfig {
	return STDOUTConfig{
		Delim: "",
	}
}

type STDOUT struct {
	writer *LineWriter
	conf   *STDOUTConfig
	inChan chan types.Chunk
}

func (stdout *STDOUT) Write(chunk types.Chunk) error {
	if stdout.writer.Closed() {
		return types.ErrOutputClosed
	}

	stdout.inChan <- chunk
	return nil
}

func (stdout *STDOUT) Close() {
	stdout.writer.CloseAsync()
}

// NewSTDOUT creates a new STDOUT output type.
func NewSTDOUT(conf Config) (Type, error) {
	writer := NewLineWriter(os.Stdout, false, []byte(conf.STDOUT.Delim))
	inChan := make(chan types.Chunk)

	// start consuming messages
	writer.Consume(inChan)

	return &STDOUT{
		writer: writer,
		conf:   &conf.STDOUT,
		inChan: inChan,
	}, nil
}
