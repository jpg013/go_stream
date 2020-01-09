package readable

type ReadableMode uint32

const (
	// ReadableFlowing mode reads from the underlying system automatically
	// and provides to an application as quickly as possible.
	ReadableFlowing ReadableMode = iota
	// ReadableNotFlowing mode is paused and data must be explicity read from the stream
	ReadableNotFlowing
	// ReadableNull mode is null, there is no mechanism for consuming the stream's data
	ReadableNull
)
