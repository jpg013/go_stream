package operators

import "github.com/jpg013/go_stream/types"

type Map struct {
	dataChan  chan types.Chunk
	errorChan chan error
	mapper    types.Mapper
}

func NewMapOperator(m types.Mapper) (types.Operator, error) {
	return &Map{
		dataChan:  make(chan types.Chunk),
		errorChan: make(chan error),
		mapper:    m,
	}, nil
}

func (m *Map) Exec(chunk types.Chunk) {
	d, err := m.mapper(chunk)

	if err != nil {
		m.errorChan <- err
	} else {
		m.dataChan <- d
	}
}

func (m *Map) GetOutput() <-chan types.Chunk {
	return m.dataChan
}

func (m *Map) GetError() <-chan error {
	return m.errorChan
}

func (m *Map) End() {
	close(m.dataChan)
	close(m.errorChan)
}
