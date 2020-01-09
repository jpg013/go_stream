package main

// import (
// 	"context"
// 	"log"

// 	"github.com/segmentio/kafka-go"
// )

// var brokers = []string{
// 	"xnd-kafka-data00.aws.dunami.com:9092",
// 	"xnd-kafka-data01.aws.dunami.com:9092",
// 	"xnd-kafka-data02.aws.dunami.com:9092",
// }

// func NewReader() *kafka.Reader {
// 	return kafka.NewReader(kafka.ReaderConfig{
// 		Brokers:     brokers,
// 		GroupID:     "go.twitter_artifact.qa",
// 		Topic:       "acquisition.twitter_artifact_raw.qa",
// 		StartOffset: kafka.FirstOffset,
// 		MinBytes:    10e3, // 10KB
// 		MaxBytes:    10e6, // 10MB
// 	})
// }

// func NewWriter() *kafka.Writer {
// 	// make a writer that produces to topic-A, using the least-bytes distribution
// 	return kafka.NewWriter(kafka.WriterConfig{
// 		Brokers:  brokers,
// 		Topic:    "acquisition.twitter_test.dev",
// 		Balancer: &kafka.LeastBytes{},
// 	})
// }

// func test() {
// 	counter := 0

// 	reader := NewReader()
// 	writer := NewWriter()

// 	msgs := make([]kafka.Message, 100)

// 	for counter < 100 {
// 		msg, err := reader.ReadMessage(context.Background())

// 		if err != nil {
// 			panic(err)
// 		}

// 		log.Printf("partition: %d", msg.Partition)
// 		log.Printf("offset: %d", msg.Offset)

// 		if err != nil {
// 			panic(err)
// 		}

// 		msg.Topic = ""
// 		msg.Partition = 0

// 		msgs[counter] = msg

// 		counter++
// 	}

// 	err := writer.WriteMessages(context.Background(), msgs...)

// 	if err != nil {
// 		panic(err)
// 	}

// 	writer.Close()
// }

// func main() {
// 	readableExample()

// 	// r := stream.FromKafka(&stream.KafkaConfig{
// 	// 	Brokers: brokers,
// 	// 	GroupID: "go_1",
// 	// 	Topic:   "acquisition.twitter_artifact_raw.dev",
// 	// })

// 	// w := stream.ToFile(&stream.FileConfig{})

// 	// The readable.pipe() method attaches a Writable stream to the readable, causing it to switch automatically into flowing mode and push all of its data to the attached Writable. The flow of data will be automatically managed so that the destination Writable stream is not overwhelmed by a faster Readable stream.
// 	// r.Pipe(w)

// 	// stream.FromKafka(&stream.KafkaConfig{
// 	// 	Brokers: brokers,
// 	// 	GroupID: "go_1",
// 	// 	Topic:   "acquisition.twitter_artifact_raw.dev",
// 	// }).Pipe(stream.ToFile(&stream.FileConfig{}))

// 	// for {
// 	// 	kafkaSource.ReadStart()

// 	// 	// lets listen for the message
// 	// 	msg := <-kafkaSource.dataC

// 	// 	y := msg.(kafka.Message)

// 	// 	// val := msg.
// 	// 	// 	fmt.Println(msg)
// 	// }
// }
