package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/brianvoe/gofakeit/v7"
	demov1 "github.com/bufbuild/bufstream-demo/gen/bufstream/demo/v1"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Producer struct {
	client     *kgo.Client
	topic      string
	serializer serde.Serializer
}

func NewProducer(
	client *kgo.Client,
	topic string,
	serializer serde.Serializer,
) *Producer {
	return &Producer{
		client:     client,
		topic:      topic,
		serializer: serializer,
	}
}

func (p *Producer) Run(ctx context.Context) (int, error) {
	cases := [][2]bool{
		{true, true},   // valid
		{true, false},  // semantically invalid
		{false, false}, // not valid protobuf
	}
	var (
		published int
		errs      []error
	)
	for _, c := range cases {
		if err := p.produce(ctx, c[0], c[1]); err != nil {
			errs = append(errs, err)
		} else {
			published++
		}
	}
	return published, errors.Join(errs...)
}

func (p *Producer) produce(ctx context.Context, validFormat, validSemantics bool) error {
	id := uuid.New().String()
	payload, err := p.payload(id, validFormat, validSemantics)
	if err != nil {
		return err
	}

	var desc string
	switch {
	case !validFormat:
		desc = "malformed message"
	case !validSemantics:
		desc = "semantically invalid message"
	default:
		desc = "valid message"
	}

	res := p.client.ProduceSync(ctx, &kgo.Record{
		Key:   []byte(id),
		Value: payload,
		Topic: p.topic,
	})
	if err := res.FirstErr(); err != nil {
		return fmt.Errorf("failed to publish: %w", err)
	}
	slog.Info("successfully published" + desc)
	return nil
}

func (p *Producer) payload(id string, validFormat, validSemantics bool) ([]byte, error) {
	if !validFormat {
		return []byte("\x00foobar"), nil
	}
	var msg *demov1.EmailUpdated
	if validSemantics {
		msg = &demov1.EmailUpdated{
			Id:              id,
			OldEmailAddress: gofakeit.Email(),
			NewEmailAddress: gofakeit.Email(),
		}
	} else {
		msg = &demov1.EmailUpdated{
			Id:              id,
			OldEmailAddress: gofakeit.Email(),
			NewEmailAddress: gofakeit.Animal(),
		}
	}
	return p.serializer.Serialize(p.topic, msg)
}
