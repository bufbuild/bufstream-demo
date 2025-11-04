// Package main implements the producer of the demo.

// The producer will example Cart messages. About 1% of messages produced
// are intentionally semantically-invalid: they contain a line with a zero
// quantity.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"sync"
	"sync/atomic"

	demov1 "github.com/bufbuild/bufstream-demo/gen/bufstream/demo/v1"
	"github.com/bufbuild/bufstream-demo/pkg/app"
	"github.com/bufbuild/bufstream-demo/pkg/kafka"
	"github.com/bufbuild/bufstream-demo/pkg/produce"
	"github.com/bufbuild/bufstream-demo/pkg/product"
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

	producer := produce.NewProducer[*demov1.Cart](
		client,
		config.Kafka.Topic,
	)

	if config.MaximumRecords != -1 {
		slog.InfoContext(ctx, fmt.Sprintf("producing %d records", config.MaximumRecords))
	} else {
		slog.InfoContext(ctx, "producing unlimited records")
	}

	var wg sync.WaitGroup
	numWorkers := 50
	attemptCount := atomic.Int64{}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				nextAttempt := attemptCount.Add(1)
				if config.MaximumRecords != -1 && nextAttempt > int64(config.MaximumRecords) {
					return
				}

				select {
				case <-ctx.Done():
					return
				default:
					var inv *demov1.Cart
					n := rand.IntN(100)
					if n < 1 {
						inv = newInvalidCart()
					} else {
						inv = newValidCart()
					}
					if err := producer.ProduceProtobufMessage(ctx, newID(), inv); err != nil {
						if errors.Is(err, context.Canceled) {
							return
						}
						slog.ErrorContext(ctx, "error producing message", "err", err)
					}
				}

				if nextAttempt > 0 && nextAttempt%250 == 0 {
					slog.InfoContext(ctx, fmt.Sprintf("produced %d records", nextAttempt))
				}
			}
		}()
	}

	wg.Wait()
	if config.MaximumRecords != -1 {
		slog.InfoContext(ctx, fmt.Sprintf("exiting after producing %d records", config.MaximumRecords))
	}
	return nil
}

// newID returns a new UUID.
//
// This is also used as the record key.
func newID() string {
	return uuid.New().String()
}

func newValidCart() *demov1.Cart {
	return newCart(
		newRandomLineItems(),
	)
}

func newInvalidCart() *demov1.Cart {
	lineItems := newRandomLineItems()
	// Invalidate a random line item by setting quantity to 0
	if len(lineItems) > 0 {
		invalidIndex := rand.IntN(len(lineItems))
		lineItems[invalidIndex].Quantity = 0
	}

	return newCart(
		lineItems,
	)
}

func newCart(lineItems []*demov1.LineItem) *demov1.Cart {
	cart := &demov1.Cart{
		CartId:    uuid.New().String(),
		LineItems: lineItems,
	}

	return cart
}

func newRandomLineItems() []*demov1.LineItem {
	maxItems := min(10, len(product.Catalog))
	numItems := rand.IntN(maxItems) + 1
	lineItems := make([]*demov1.LineItem, 0, numItems)

	// Track product_ids to ensure uniqueness.
	usedProductIDs := make(map[string]bool)

	for len(lineItems) < numItems {
		randomProduct := randomProduct()

		// Skip if we've already used this randomProduct.
		if usedProductIDs[randomProduct.GetProductId()] {
			continue
		}

		usedProductIDs[randomProduct.GetProductId()] = true

		lineItems = append(lineItems, &demov1.LineItem{
			LineItemId:     uuid.New().String(),
			Product:        randomProduct,
			Quantity:       uint64(rand.IntN(5) + 1),
			UnitPriceCents: randomProduct.GetUnitPriceCents(),
		})
	}

	return lineItems
}

// RandomProduct returns a randomly selected product from the catalog.
func randomProduct() *demov1.Product {
	return product.Catalog[rand.IntN(len(product.Catalog))]
}
