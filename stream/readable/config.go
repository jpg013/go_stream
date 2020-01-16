package readable

import "github.com/jpg013/go_stream/types"

// Config represents a readable config
type Config struct {
	highWaterMark int
	generator     types.Generator
}

// OptionFunc applies defaults to a readable config
type OptionFunc func(*Config)

// NewConfig returns a new readable config. It uses option
// functions to provide config defaults.
func NewConfig(fns ...OptionFunc) *Config {
	conf := &Config{}

	for _, fn := range fns {
		fn(conf)
	}

	return conf
}
