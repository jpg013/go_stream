package writable

import "github.com/jpg013/go_stream/output"

// Config represents a writable config
type Config struct {
	HighWaterMark int
	Out           output.Type
}

// OptionFunc applies defaults to a writable config
type OptionFunc func(*Config)

// NewConfig returns a new writable config. It uses option
// functions to provide config defaults.
func NewConfig(fns ...OptionFunc) *Config {
	conf := &Config{}

	for _, fn := range fns {
		fn(conf)
	}

	return conf
}
