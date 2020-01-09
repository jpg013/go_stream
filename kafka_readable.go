package main

// import (
// 	"github.com/segmentio/kafka-go"
// )

// type KafkaConfig struct {
// 	Brokers []string
// 	GroupID string
// 	Topic   string
// }

// type KafkaReadable struct {
// 	reader     *kafka.Reader
// 	state      *ReadableState
// 	streamType StreamType
// }

// func (k *KafkaReadable) ReadStart() (interface{}, error) {
// 	if k.isReading == false {
// 		k.Open()
// 	}

// 	msg := <-k.messageQueue

// 	return msg, nil
// }

// func (k *KafkaReadable) ReadStop() error {
// 	k.mux.Lock()
// 	defer k.mux.Unlock()

// 	// set isReading to false
// 	k.isReading = false

// 	// close message channel
// 	close(k.messageQueue)

// 	k.reader.Close()

// 	return nil
// }

// func (k *KafkaReadable) Open() error {
// 	k.mux.Lock()
// 	defer k.mux.Unlock()

// 	if k.isReading {
// 		return nil
// 	}

// 	k.messageQueue = make(chan kafka.Message)

// 	go func() {
// 		for k.isReading {
// 			msg, err := k.reader.ReadMessage(context.Background())

// 			if err != nil {
// 				k.errorCh <- err
// 				k.ReadStop()
// 				return
// 			}

// 			k.messageQueue <- msg
// 		}
// 	}()

// 	k.isReading = true

// 	return nil
// }

// func (k *KafkaReadable) Pipe(w Writable) Stream {
// 	// k.pipes = append()
// }

// func (k *KafkaReadable) GetState() *ReadableState {
// 	return k.state
// }

// func (k *KafkaReadable) Read() {
// 	msg, _ := k.reader.ReadMessage(context.Background())

// 	readableAddChunk(k, msg)
// }

// func FromKafka(settings *KafkaConfig) Readable {
// 	options := kafka.ReaderConfig{
// 		Brokers:     settings.Brokers,
// 		GroupID:     settings.GroupID,
// 		Topic:       settings.Topic,
// 		StartOffset: kafka.FirstOffset,
// 		MinBytes:    10e3, // 10KB
// 		MaxBytes:    10e6, // 10MB
// 	}

// 	return &KafkaReadable{
// 		state:      NewReadableState(),
// 		streamType: ReadableType,
// 		reader:     kafka.NewReader(options),
// 	}
// }
