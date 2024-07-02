package main

import (
	"context"
	"fmt"
	"log/slog"

	bufstream_demov1 "github.com/bufbuild/bufstream-demo/gen/bufbuild/bufstream_demo/v1"
	"github.com/bufbuild/bufstream-demo/storage"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Verifier struct {
	store        storage.Interface
	topic        string
	deserializer serde.Deserializer
	consumer     *kgo.Client
}

func NewVerifier(
	store storage.Interface,
	topic string,
	deserializer serde.Deserializer,
	consumer *kgo.Client,
) *Verifier {
	return &Verifier{
		store:        store,
		topic:        topic,
		deserializer: deserializer,
		consumer:     consumer,
	}
}

func (d *Verifier) Run(ctx context.Context) (retErr error) {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := d.consume(ctx); err != nil {
				return err
			}
		}
	}
}

func (d *Verifier) consume(ctx context.Context) error {
	fetches := d.consumer.PollFetches(ctx)
	if errs := fetches.Errors(); len(errs) > 0 {
		return fmt.Errorf("failed to poll fetches: %v", errs)
	}
	for _, record := range fetches.Records() {
		msg, err := d.deserializer.Deserialize(record.Topic, record.Value)
		if err != nil {
			return err
		}
		updated, ok := msg.(*bufstream_demov1.EmailUpdated)
		if !ok {
			slog.WarnContext(ctx, "unexpected message received", "msg", msg)
		} else if err = d.store.VerifyEmail(ctx, updated.GetUuid(), updated.GetNewAddress()); err != nil {
			slog.WarnContext(ctx, "failed to verify email",
				"uuid", updated.GetUuid(), "error", err)
		} else {
			slog.InfoContext(ctx, "verified email",
				"uuid", updated.GetUuid(),
				"old_address", updated.GetOldAddress(),
				"new_address", updated.GetNewAddress(),
			)
		}
	}
	return nil
}
