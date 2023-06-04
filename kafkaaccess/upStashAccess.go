package kafkaaccess

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

func ConnectAsProdcer() {
	mchanism, err := scram.Mechanism(scram.SHA256, "YWNlLWdvcGhlci0xMTM5NyTV8OvXnNgwFjsXuvxTrqggb2zdNFk1IqhkvLa2sEs", "75406d5d46d94844b751b2b458d9999e")

	dialer := &kafka.Dialer{
		SASLMechanism: mchanism,
		TLS:           &tls.Config{},
	}

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"ace-gopher-11397-eu1-kafka.upstash.io:9092"},
		Topic:   "User",
		Dialer:  dialer,
	})
	//...
	defer w.Close()

	err = w.WriteMessages(context.Background(),
		kafka.Message{
			Value: []byte("Hello Upstash!"),
		},
	)
	if err != nil {
		log.Fatalln("failed to write messages:", err)
	}
}

func ConnectAsConsumer() {
	mechanism, err := scram.Mechanism(scram.SHA256, "YWNlLWdvcGhlci0xMTM5NyTV8OvXnNgwFjsXuvxTrqggb2zdNFk1IqhkvLa2sEs", "75406d5d46d94844b751b2b458d9999e")
	if err != nil {
		log.Fatalln(err)
	}

	dialer := &kafka.Dialer{
		SASLMechanism: mechanism,
		TLS:           &tls.Config{},
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"ace-gopher-11397-eu1-kafka.upstash.io:9092"},
		GroupID: "GROUP_NAME",
		Topic:   "User",
		Dialer:  dialer,
	})
	defer r.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1000)
	defer cancel()

	m, err := r.ReadMessage(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("%+v\n", m)
}
