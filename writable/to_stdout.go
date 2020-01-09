package writable

import "stream/output"

func ToSTDOUT() (Writable, error) {
	output, err := output.NewSTDOUT(output.NewConfig())

	if err != nil {
		return nil, err
	}

	return NewWritable(output), nil
}
