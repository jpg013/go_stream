package output

// Config is the all encompassing configuration struct for all output types.
type Config struct {
	File   FileConfig   `json:"file"`
	STDOUT STDOUTConfig `json:"stdout"`
}

// NewConfig returns a configuration struct fully populated with default values.
func NewConfig() Config {
	return Config{
		File:   NewFileConfig(),
		STDOUT: NewSTDOUTConfig(),
	}
}
