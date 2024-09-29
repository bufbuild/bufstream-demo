package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	demov1 "github.com/bufbuild/bufstream-demo/gen/bufstream/demo/v1"
	"github.com/bufbuild/bufstream-demo/pkg/app"
	"github.com/bufbuild/bufstream-demo/pkg/consume"
	"github.com/bufbuild/bufstream-demo/pkg/csr"
	"github.com/bufbuild/bufstream-demo/pkg/kafka"
	"github.com/bufbuild/bufstream-demo/pkg/produce"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
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

	var serializer serde.Serializer
	var deserializer serde.Deserializer
	if config.CSR.URL != "" {
		csrClient, err := csr.NewCSRClient(config.CSR)
		if err != nil {
			return err
		}
		serializer, err = csr.NewCSRProtobufSerializer(csrClient)
		if err != nil {
			return err
		}
		deserializer, err = csr.NewCSRProtobufDeserializer(csrClient)
		if err != nil {
			return err
		}
	} else {
		serializer = csr.NewSingleTypeProtobufSerializer[*demov1.EmailUpdated]()
		deserializer = csr.NewSingleTypeProtobufDeserializer[*demov1.EmailUpdated]()
	}
	defer func() { _ = serializer.Close() }()
	defer func() { _ = deserializer.Close() }()

	producer := produce.NewProducer[*demov1.EmailUpdated](
		client,
		serializer,
		config.Kafka.Topic,
	)
	consumer := consume.NewConsumer[*demov1.EmailUpdated](
		client,
		deserializer,
		config.Kafka.Topic,
		consume.WithMessageHandler(handleEmailUpdated),
	)

	for {
		id := newID()
		if err := producer.ProduceProtobufMessage(ctx, id, newSemanticallyValidEmailUpdated(id)); err != nil {
			return err
		}
		id = newID()
		if err := producer.ProduceProtobufMessage(ctx, id, newSemanticallyInvalidEmailUpdated(id)); err != nil {
			return err
		}
		id = newID()
		if err := producer.ProduceInvalid(ctx, id); err != nil {
			return err
		}
		if err := consumer.Consume(ctx, 3); err != nil {
			return err
		}
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

// TODO: Maybe just remove this entirely.
func handleEmailUpdated(message *demov1.EmailUpdated) error {
	var suffix string
	if old := message.GetOldEmailAddress(); old == "" {
		suffix = "redacted old email"
	} else {
		suffix = fmt.Sprintf("old email %s", old)
	}
	slog.Info(fmt.Sprintf("received message with new email %s and %s", message.GetNewEmailAddress(), suffix))
	return nil
}
