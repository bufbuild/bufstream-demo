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

	// NewSerializer creates a CSR-based Serializer if there is a CSR URL,
	// otherwise it creates a single-type Serializer for demov1.EmailUpdated.
	//
	// If a CSR URL is provided, the data will be enveloped when sent to Bufstream.
	// If not, the data will be unenveloped, however Bufstream is schema-aware, and
	// has the capability to automatically envelope data if the "coerce" configuration
	// setting is set. See the documentation for more details.
	serializer, err := csr.NewSerializer[*demov1.EmailUpdated](config.CSR)
	if err != nil {
		return err
	}
	defer func() { _ = serializer.Close() }()

	producer := produce.NewProducer[*demov1.EmailUpdated](
		client,
		serializer,
		config.Kafka.Topic,
	)

	slog.Info("starting produce")
	for {
		id := newID()
		// Produces semantically valid EmailUpdated message, where both email
		// fields are valid email addresses.
		if err := producer.ProduceProtobufMessage(ctx, id, newSemanticallyValidEmailUpdated(id)); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			slog.Error("error on produce of semantically valid protobuf message", "error", err)
		} else {
			slog.Info("produced semantically valid protobuf message", "id", id)
		}
		id = newID()
		// Produces a semantically invalid EmailUpdated message, where the new email field
		// is not a valid email address.
		if err := producer.ProduceProtobufMessage(ctx, id, newSemanticallyInvalidEmailUpdated(id)); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			slog.Error("error on produce of semantically invalid protobuf message", "error", err)
		} else {
			slog.Info("produced semantically invalid protobuf message", "id", id)
		}
		id = newID()
		// Produces a record containing a payload that is not valid Protobuf.
		if err := producer.ProduceInvalid(ctx, id); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			slog.Error("error on produce of invalid data", "error", err)
		} else {
			slog.Info("produced invalid data", "id", id)
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
