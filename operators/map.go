package operators

import "github.com/jpg013/go_stream/types"

type Map struct {
	dataChan  chan types.Chunk
	errorChan chan error
	mapper    Mapper
}

func NewMapOperator(m Mapper) (Type, error) {
	return &Map{
		dataChan:  make(chan types.Chunk),
		errorChan: make(chan error),
		mapper:    m,
	}, nil
}

func (m *Map) Exec(chunk types.Chunk) error {
	d, err := m.mapper(chunk)

	if err != nil {
		m.errorChan <- err
	} else {
		m.dataChan <- d
	}

	return nil
}

func (m *Map) Output() <-chan types.Chunk {
	return m.dataChan
}

func (m *Map) Error() <-chan error {
	return m.errorChan
}

func (m *Map) End() {
	close(m.dataChan)
	close(m.errorChan)
}
