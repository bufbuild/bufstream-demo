// Package main implements the consumer of the demo.
//
// The consumer will read as many records it can at once, print what it received,
// and then loop.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	demov1 "github.com/bufbuild/bufstream-demo/gen/bufstream/demo/v1"
	"github.com/bufbuild/bufstream-demo/pkg/app"
	"github.com/bufbuild/bufstream-demo/pkg/consume"
	"github.com/bufbuild/bufstream-demo/pkg/kafka"
)

func main() {
	// See the app package for the boilerplate we use to set up the producer and
	// consumer, including bound flags.
	app.Main(run)
}

var cartsHandled = 0

func run(ctx context.Context, config app.Config) error {
	client, err := kafka.NewKafkaClient(config.Kafka, true)
	if err != nil {
		return err
	}
	defer client.Close()

	consumer := consume.NewConsumer(
		client,
		config.Kafka.Topic,
		consume.WithMessageHandler(handleCart),
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

func handleCart(_ context.Context, invoice *demov1.Cart) error {
	for _, lineItem := range invoice.GetLineItems() {
		if lineItem.GetQuantity() == 0 {
			slog.Error("received a Cart with a zero-quantity LineItem")
		}
	}

	cartsHandled++
	if cartsHandled%250 == 0 {
		slog.Info(fmt.Sprintf("received %d carts", cartsHandled))
	}

	return nil
}
