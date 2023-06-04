package kafkaaccess

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

func ConnectAndConsumeMessage() {
	mechanism, err := scram.Mechanism(scram.SHA512,
		"YWNlLWdvcGhlci0xMTM5NyTV8OvXnNgwFjsXuvxTrqggb2zdNFk1IqhkvLa2sEs", "75406d5d46d94844b751b2b458d9999e")
	if err != nil {
		log.Fatalln(err)
	}

	dialer := &kafka.Dialer{
		SASLMechanism: mechanism,
		TLS:           &tls.Config{},
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"ace-gopher-11397-eu1-kafka.upstash.io:9092"},
		Topic:   "User",
		Dialer:  dialer,
	})
	defer r.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			log.Fatalln(err)
		}
		msg := string(m.Value)
		log.Printf("%+v\n", msg)
	}
}
