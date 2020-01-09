package writable

// import (
// 	"bufio"
// 	"bytes"
// 	"encoding/gob"
// 	"os"
// 	"stream/types"
// )

// type FileWriterConfig struct {
// 	Filename  string
// 	FlushSize int
// }

// type FileWriter struct {
// 	f      *os.File
// 	buffer *bufio.Writer
// 	conf   *FileWriterConfig
// }

// func (w *FileWriter) End() {
// 	if w.buffer.Size() > 0 {
// 		w.buffer.Flush()
// 	}
// }

// // https://stackoverflow.com/questions/23003793/convert-arbitrary-golang-interface-to-byte-array
// func getBytes(data types.Chunk) ([]byte, error) {
// 	var buf bytes.Buffer
// 	enc := gob.NewEncoder(&buf)
// 	err := enc.Encode(data)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return buf.Bytes(), nil
// }

// func (w *FileWriter) DoWrite(data types.Chunk) error {
// 	b, err := getBytes(data)
// 	if err != nil {
// 		return err
// 	}

// 	w.buffer.Write(b)

// 	if w.buffer.Size() >= w.conf.FlushSize {
// 		w.buffer.Flush()
// 	}

// 	return nil
// }

// func fileExists(filename string) bool {
// 	info, err := os.Stat(filename)

// 	if os.IsNotExist(err) {
// 		return false
// 	}

// 	return !info.IsDir()
// }

// func openOrCreateFile(filename string) (*os.File, error) {
// 	if fileExists(filename) {
// 		return os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0666)
// 	}

// 	return os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
// }

// func ToFile(conf *FileWriterConfig) Writable {
// 	// open or create file handle
// 	f, err := openOrCreateFile(conf.Filename)

// 	if err != nil {
// 		panic(err.Error())
// 	}

// 	writer := &FileWriter{
// 		f:      f,
// 		buffer: bufio.NewWriter(f),
// 		conf:   conf,
// 	}

// 	return NewWritable(writer)
// }
