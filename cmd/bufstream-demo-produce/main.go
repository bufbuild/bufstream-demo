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

	// Producers for browsing events
	searchProducer := produce.NewProducer[*demov1.ProductsSearched](client, config.SearchTopic)
	listViewedProducer := produce.NewProducer[*demov1.ProductListViewed](client, config.ListViewedTopic)
	listFilteredProducer := produce.NewProducer[*demov1.ProductListFiltered](client, config.ListFilteredTopic)
	// Producer for email updates
	emailProducer := produce.NewProducer[*demov1.EmailUpdated](client, config.Kafka.Topic)

	slog.Info("starting produce")
	for {
		id := newID()
		// Produce ProductsSearched event
		if config.SearchTopic != "" {
			if err := searchProducer.ProduceProtobufMessage(ctx, id, newProductsSearched(id)); err != nil {
				slog.Error("error producing ProductsSearched event", "error", err)
			} else {
				slog.Info("produced ProductsSearched event", "id", id)
			}
		}
		// Produce ProductListViewed event
		id = newID()
		if config.ListViewedTopic != "" {
			if err := listViewedProducer.ProduceProtobufMessage(ctx, id, newProductListViewed(id)); err != nil {
				slog.Error("error producing ProductListViewed event", "error", err)
			} else {
				slog.Info("produced ProductListViewed event", "id", id)
			}
		}
		// Produce ProductListFiltered event
		id = newID()
		if config.ListFilteredTopic != "" {
			if err := listFilteredProducer.ProduceProtobufMessage(ctx, id, newProductListFiltered(id)); err != nil {
				slog.Error("error producing ProductListFiltered event", "error", err)
			} else {
				slog.Info("produced ProductListFiltered event", "id", id)
			}
		}
		// Produces semantically valid EmailUpdated message
		id = newID()
		if err := emailProducer.ProduceProtobufMessage(ctx, id, newSemanticallyValidEmailUpdated(id)); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			slog.Error("error on produce of semantically valid protobuf message", "error", err)
		} else {
			slog.Info("produced semantically valid protobuf message", "id", id)
		}
		// Produces a semantically invalid EmailUpdated message, where the new email field
		// is not a valid email address.
		id = newID()
		if err := emailProducer.ProduceProtobufMessage(ctx, id, newSemanticallyInvalidEmailUpdated(id)); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			slog.Error("error on produce of semantically invalid protobuf message", "error", err)
		} else {
			slog.Info("produced semantically invalid protobuf message", "id", id)
		}
		id = newID()
		// Produces a record containing a payload that is not valid Protobuf.
		if err := emailProducer.ProduceInvalid(ctx, id); err != nil {
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

// newProductsSearched generates a dummy ProductsSearched event
func newProductsSearched(id string) *demov1.ProductsSearched {
	return &demov1.ProductsSearched{
		Id:    id,
		Query: gofakeit.Word(),
	}
}

// newProductListViewed generates a dummy ProductListViewed event
func newProductListViewed(id string) *demov1.ProductListViewed {
	products := make([]*demov1.Product, gofakeit.Number(1, 5))
	for i := range products {
		products[i] = &demov1.Product{
			ProductId: gofakeit.UUID(),
			Sku:       gofakeit.Word(),
			Name:      gofakeit.ProductName(),
			Price:     float64(gofakeit.Price(1, 100)),
			Position:  int32(i + 1),
			Category:  gofakeit.Word(),
			Url:       gofakeit.URL(),
			ImageUrl:  gofakeit.URL(),
		}
	}
	return &demov1.ProductListViewed{
		Id:       id,
		ListId:   gofakeit.UUID(),
		Category: gofakeit.Word(),
		Products: products,
	}
}

// newProductListFiltered generates a dummy ProductListFiltered event
func newProductListFiltered(id string) *demov1.ProductListFiltered {
	filters := []*demov1.Filter{{Type: "department", Value: gofakeit.Word()}}
	sorts := []*demov1.Sort{{Type: "price", Value: "asc"}}
	products := make([]*demov1.Product, gofakeit.Number(1, 5))
	for i := range products {
		products[i] = &demov1.Product{
			ProductId: gofakeit.UUID(),
			Sku:       gofakeit.Word(),
			Name:      gofakeit.ProductName(),
			Price:     float64(gofakeit.Price(1, 100)),
			Position:  int32(i + 1),
			Category:  gofakeit.Word(),
			Url:       gofakeit.URL(),
			ImageUrl:  gofakeit.URL(),
		}
	}
	return &demov1.ProductListFiltered{
		Id:       id,
		ListId:   gofakeit.UUID(),
		Filters:  filters,
		Sorts:    sorts,
		Products: products,
	}
}
