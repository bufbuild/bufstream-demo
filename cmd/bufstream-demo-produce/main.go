// Package main implements the producer of the demo.

// This is run as part of docker compose.
//
// The producer will produce three records at a time:
//
//   - A semantically-valid EmailUpdated message, where both email fields are valid email addresses.
//   - A semantically-invalid EmailUpdated message, where the new email field is not a valid email address.
//   - A record containing a payload that is not valid Protobuf.
//
// After producing these three records, the producer sleeps for one second, and then loops.
package main

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	demov1 "github.com/bufbuild/bufstream-demo/gen/bufstream/demo/v1"
	"github.com/bufbuild/bufstream-demo/pkg/app"
	"github.com/bufbuild/bufstream-demo/pkg/csr"
	"github.com/bufbuild/bufstream-demo/pkg/kafka"
	"github.com/bufbuild/bufstream-demo/pkg/produce"
	"github.com/google/uuid"
)

func main() {
	// See the app package for the boilerplate we use to set up the producer and
	// consumer, including bound flags.
	app.MainAutoCreateTopic(run)
}

func run(ctx context.Context, config app.Config) error {
	client, err := kafka.NewKafkaClient(config.Kafka, false)
	if err != nil {
		return err
	}
	defer client.Close()

	// NewSerde creates a CSR-based serializer if there is a CSR URL,
	// otherwise it creates a single-type serializer for demov1.EmailUpdated.
	serde, err := csr.NewSerde[*demov1.EmailUpdated](ctx, config.CSR, config.Kafka.Topic)
	if err != nil {
		return err
	}

	producer := produce.NewProducer[*demov1.EmailUpdated](
		client,
		serde,
		config.Kafka.Topic,
	)

	slog.InfoContext(ctx, "starting produce")
	for {
		msgID := newID()
		// Produces semantically-valid EmailUpdated message, where both email
		// fields are valid email addresses.
		if err := producer.ProduceProtobufMessage(ctx, msgID, newSemanticallyValidEmailUpdated(msgID)); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			slog.ErrorContext(ctx, "error on produce of semantically valid protobuf message", "error", err)
		} else {
			slog.InfoContext(ctx, "produced semantically valid protobuf message", "id", msgID)
		}
		// TODO: remove this short circuit. It's testing one valid message.
		return nil

		msgID = newID()
		// Produces a semantically-invalid EmailUpdated message, where the new email field
		// is not a valid email address.
		if err := producer.ProduceProtobufMessage(ctx, msgID, newSemanticallyInvalidEmailUpdated(msgID)); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			slog.ErrorContext(ctx, "error on produce of semantically invalid protobuf message", "error", err)
		} else {
			slog.InfoContext(ctx, "produced semantically invalid protobuf message", "id", msgID)
		}
		msgID = newID()
		// Produces record containing a payload that is not valid Protobuf.
		if err := producer.ProduceInvalid(ctx, msgID); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			slog.ErrorContext(ctx, "error on produce of invalid data", "error", err)
		} else {
			slog.InfoContext(ctx, "produced invalid data", "id", msgID)
		}
		time.Sleep(time.Second)
	}
}

// newID returns a new UUID.
//
// This is also used as the record key.
func newID() string {
	return uuid.New().String()
}

func newSemanticallyValidEmailUpdated(id string) *demov1.EmailUpdated {
	return &demov1.EmailUpdated{
		Id:              id,
		OldEmailAddress: gofakeit.Email(),
		NewEmailAddress: gofakeit.Email(),
	}
}

func newSemanticallyInvalidEmailUpdated(id string) *demov1.EmailUpdated {
	return &demov1.EmailUpdated{
		Id:              id,
		OldEmailAddress: gofakeit.Email(),
		NewEmailAddress: gofakeit.Animal(),
	}
}
