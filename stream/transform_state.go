package stream

type TransformState struct {
	transforming uint32
}

func NewTransformState() *TransformState {
	return &TransformState{
		transforming: 0,
	}
}
