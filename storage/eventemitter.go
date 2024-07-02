package storage

import (
	"context"
	"log/slog"

	"github.com/bufbuild/bufstream-demo/gen/bufbuild/bufstream_demo/v1"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/twmb/franz-go/pkg/kgo"
)

type EventEmitter struct {
	inner      Interface
	topic      string
	serializer serde.Serializer
	producer   *kgo.Client
}

func NewEventEmitter(inner Interface, topic string, serializer serde.Serializer, producer *kgo.Client) *EventEmitter {
	return &EventEmitter{
		inner:      inner,
		topic:      topic,
		serializer: serializer,
		producer:   producer,
	}
}

func (em EventEmitter) GetEmail(ctx context.Context, uuid string) (*bufstream_demov1.UserEmail, error) {
	return em.inner.GetEmail(ctx, uuid)
}

func (em EventEmitter) VerifyEmail(ctx context.Context, uuid, address string) error {
	return em.inner.VerifyEmail(ctx, uuid, address)
}

func (em EventEmitter) UpdateEmail(ctx context.Context, uuid, address string) (previous string, err error) {
	previous, err = em.inner.UpdateEmail(ctx, uuid, address)
	if err != nil {
		return previous, err
	}
	evt := &bufstream_demov1.EmailUpdated{
		Uuid:       uuid,
		OldAddress: previous,
		NewAddress: address,
	}
	payload, err := em.serializer.Serialize(em.topic, evt)
	if err != nil {
		slog.WarnContext(ctx, "failed to serialize event", "error", err)
		return previous, nil
	}
	res := em.producer.ProduceSync(ctx, &kgo.Record{
		Topic: em.topic,
		Key:   []byte(uuid),
		Value: payload,
	})
	if err := res.FirstErr(); err != nil {
		slog.WarnContext(ctx, "failed to produce event", "error", err)
	} else {
		slog.InfoContext(ctx, "email updated event produced", "uuid", uuid, "address", address)
	}
	return previous, nil
}
