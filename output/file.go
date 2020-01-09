package output

import (
	"github.com/jpg013/go_stream/types"
	"os"
)

// FileConfig contains configuration fields for the file based output type.
type FileConfig struct {
	Path  string `json:"path"`
	Delim string `json:"delimiter"`
}

// NewFileConfig creates a new FileConfig with default values.
func NewFileConfig() FileConfig {
	return FileConfig{
		Path:  "",
		Delim: "",
	}
}

type File struct {
	writer *LineWriter
	conf   *FileConfig
	inChan chan types.Chunk
}

func (f *File) Write(chunk types.Chunk) error {
	if f.writer.Closed() {
		return types.ErrOutputClosed
	}

	f.inChan <- chunk
	return nil
}

func (f *File) Close() {
	f.writer.CloseAsync()
}

// NewFile creates a new File output type.
func NewFile(conf Config) (Type, error) {
	file, err := os.OpenFile(conf.File.Path, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.FileMode(0666))
	if err != nil {
		return nil, err
	}

	writer := NewLineWriter(file, false, []byte(conf.File.Delim))
	inChan := make(chan types.Chunk)

	// start consuming messages
	writer.Consume(inChan)

	return &File{
		writer: writer,
		conf:   &conf.File,
		inChan: inChan,
	}, nil
}
