package main

import (
	"context"
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
	app.Main(run)
}

func run(ctx context.Context, config app.Config) error {
	client, err := kafka.NewKafkaClient(config.Kafka)
	if err != nil {
		return err
	}
	defer client.Close()

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

	for {
		id := newID()
		if err := producer.ProduceProtobufMessage(ctx, id, newSemanticallyValidEmailUpdated(id)); err != nil {
			return err
		}
		slog.Info("produced semantically valid protobuf message", "id", id)
		id = newID()
		if err := producer.ProduceProtobufMessage(ctx, id, newSemanticallyInvalidEmailUpdated(id)); err != nil {
			return err
		}
		slog.Info("produced semantically invalid protobuf message", "id", id)
		id = newID()
		if err := producer.ProduceInvalid(ctx, id); err != nil {
			return err
		}
		slog.Info("produced invalid data", "id", id)
		time.Sleep(time.Second)
	}
}

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
