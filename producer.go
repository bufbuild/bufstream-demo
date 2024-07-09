package main

import (
	"context"
	"errors"
	"log/slog"

	"github.com/brianvoe/gofakeit/v7"
	bufstream_demov1 "github.com/bufbuild/bufstream-demo/gen/bufbuild/bufstream_demo/v1"
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

func (p *Producer) Run(ctx context.Context) error {
	return errors.Join(
		p.produce(ctx, true, true),
		p.produce(ctx, true, false),
		p.produce(ctx, false, false),
	)
}

func (p *Producer) produce(ctx context.Context, validFormat, validSemantics bool) error {
	id := uuid.New().String()
	payload, err := p.payload(id, validFormat, validSemantics)
	if err != nil {
		return err
	}
	format := "valid"
	switch {
	case !validFormat:
		format = "malformed"
	case !validSemantics:
		format = "invalid"
	}

	logger := slog.With("format", format)
	res := p.client.ProduceSync(ctx, &kgo.Record{
		Key:   []byte(id),
		Value: payload,
		Topic: p.topic,
	})
	if err = res.FirstErr(); err != nil {
		logger.Error("failed to produce record", "error", err)
	} else {
		logger.Info("produced record")
	}
	return nil
}

func (p *Producer) payload(id string, validFormat, validSemantics bool) ([]byte, error) {
	if !validFormat {
		return []byte("\x00foobar"), nil
	}
	var msg *bufstream_demov1.EmailUpdated
	if validSemantics {
		msg = &bufstream_demov1.EmailUpdated{
			Uuid:       id,
			OldAddress: gofakeit.Email(),
			NewAddress: gofakeit.Email(),
		}
	} else {
		msg = &bufstream_demov1.EmailUpdated{
			Uuid:       id,
			OldAddress: gofakeit.Email(),
			NewAddress: gofakeit.Animal(),
		}
	}
	return p.serializer.Serialize(p.topic, msg)
}
