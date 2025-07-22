// Package main implements the consumer of the demo.
//
// This is run as part of docker compose.
//
// The consumer will read as many records it can at once, print what it received,
// sleep for one second, and then loop.
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
	// See the app package for the boilerplate we use to set up the producer and
	// consumer, including bound flags.
	app.Main(run)
}

func run(ctx context.Context, config app.Config) error {
	client, err := kafka.NewKafkaClient(config.Kafka, true)
	if err != nil {
		return err
	}
	defer client.Close()

	// NewSerde creates a CSR-based deserializer if there is a CSR URL,
	// otherwise it creates a single-type deserializer for demov1.EmailUpdated.
	serde, err := csr.NewSerde[*demov1.EmailUpdated](ctx, config.CSR, config.Kafka.Topic)
	if err != nil {
		return err
	}

	consumer := consume.NewConsumer(
		client,
		serde,
		config.Kafka.Topic,
		consume.WithMessageHandler(handleEmailUpdated),
	)

	slog.InfoContext(ctx, "starting consume")
	for {
		// Read as many messages as we can.
		//
		// Only return error if there is an unexpected system error. Of note, an error is not
		// returned if the data that the consumer receives is malformed.
		if err := consumer.Consume(ctx); err != nil {
			return err
		}
		time.Sleep(time.Second)
	}
}

func handleEmailUpdated(ctx context.Context, message *demov1.EmailUpdated) error {
	var suffix string
	if old := message.GetOldEmailAddress(); old == "" {
		suffix = "redacted old email"
	} else {
		suffix = "old email " + old
	}
	slog.InfoContext(ctx, fmt.Sprintf("consumed message with new email %s and %s", message.GetNewEmailAddress(), suffix))
	return nil
}
