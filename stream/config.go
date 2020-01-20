package stream

import (
	"github.com/jpg013/go_stream/output"
	"github.com/jpg013/go_stream/types"
)

// OptionFunc applies defaults to a config
type OptionFunc func(*Config)

type TransformConfig struct {
	Operator types.Operator
}

type ReadableConfig struct {
	HighWaterMark int
	Generator     types.Generator
}

// Config represents a readable config
type Config struct {
	Transform *TransformConfig
	Readable  *ReadableConfig
	Writable  *WritableConfig
}

// NewConfig returns a new stream config. It uses option
// functions to provide config defaults.
func NewConfig(fns ...OptionFunc) *Config {
	conf := &Config{
		Transform: &TransformConfig{},
		Readable:  &ReadableConfig{},
		Writable:  &WritableConfig{},
	}

	for _, fn := range fns {
		fn(conf)
	}

	return conf
}

// Config represents a writable config
type WritableConfig struct {
	HighWaterMark int
	Out           output.Type
}
