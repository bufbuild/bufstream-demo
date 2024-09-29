package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	demov1 "github.com/bufbuild/bufstream-demo/gen/bufstream/demo/v1"
	"github.com/bufbuild/bufstream-demo/pkg/app"
	"github.com/bufbuild/bufstream-demo/pkg/consume"
	"github.com/bufbuild/bufstream-demo/pkg/csr"
	"github.com/bufbuild/bufstream-demo/pkg/kafka"
)

func main() {
	app.Main(run)
}

func run(ctx context.Context, config app.Config) error {
	client, err := kafka.NewKafkaClient(config.Kafka)
	if err != nil {
		return err
	}
	defer client.Close()

	deserializer, err := csr.NewDeserializer[*demov1.EmailUpdated](config.CSR)
	if err != nil {
		return err
	}
	defer func() { _ = deserializer.Close() }()

	consumer := consume.NewConsumer[*demov1.EmailUpdated](
		client,
		deserializer,
		config.Kafka.Topic,
		consume.WithMessageHandler(handleEmailUpdated),
	)

	for {
		if err := consumer.Consume(ctx); err != nil {
			return err
		}
		time.Sleep(time.Second)
	}
}

// TODO: Maybe just remove this entirely.
func handleEmailUpdated(message *demov1.EmailUpdated) error {
	var suffix string
	if old := message.GetOldEmailAddress(); old == "" {
		suffix = "redacted old email"
	} else {
		suffix = fmt.Sprintf("old email %s", old)
	}
	slog.Info(fmt.Sprintf("consumed message with new email %s and %s", message.GetNewEmailAddress(), suffix))
	return nil
}
