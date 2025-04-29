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
	demov1 "github.com/bufbuild/bufstream-demo/gen/bufstream/demo/v1" // Corrected import path
	"github.com/bufbuild/bufstream-demo/pkg/app"
	"github.com/bufbuild/bufstream-demo/pkg/kafka"
	"github.com/bufbuild/bufstream-demo/pkg/produce"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
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
	// Producers for ordering events
	productClickedProducer := produce.NewProducer[*demov1.ProductClicked](client, config.ProductClickedTopic)
	productViewedProducer := produce.NewProducer[*demov1.ProductViewed](client, config.ProductViewedTopic)
	productAddedProducer := produce.NewProducer[*demov1.ProductAdded](client, config.ProductAddedTopic)
	productRemovedProducer := produce.NewProducer[*demov1.ProductRemoved](client, config.ProductRemovedTopic)
	cartViewedProducer := produce.NewProducer[*demov1.CartViewed](client, config.CartViewedTopic)
	checkoutStartedProducer := produce.NewProducer[*demov1.CheckoutStarted](client, config.CheckoutStartedTopic)
	checkoutStepViewedProducer := produce.NewProducer[*demov1.CheckoutStepViewed](client, config.CheckoutStepViewedTopic)
	checkoutStepCompletedProducer := produce.NewProducer[*demov1.CheckoutStepCompleted](client, config.CheckoutStepCompletedTopic)
	paymentInfoEnteredProducer := produce.NewProducer[*demov1.PaymentInfoEntered](client, config.PaymentInfoEnteredTopic)
	orderUpdatedProducer := produce.NewProducer[*demov1.OrderUpdated](client, config.OrderUpdatedTopic)
	orderCompletedProducer := produce.NewProducer[*demov1.OrderCompleted](client, config.OrderCompletedTopic)
	orderRefundedProducer := produce.NewProducer[*demov1.OrderRefunded](client, config.OrderRefundedTopic)
	orderCancelledProducer := produce.NewProducer[*demov1.OrderCancelled](client, config.OrderCancelledTopic)

	slog.Info("waiting a few seconds for broker to stabilize after topic creation...")
	time.Sleep(5 * time.Second) // Add a delay

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

		// Produce Ordering events
		produceOrderingEvent(ctx, config.ProductClickedTopic, productClickedProducer, newProductClicked)
		produceOrderingEvent(ctx, config.ProductViewedTopic, productViewedProducer, newProductViewed)
		produceOrderingEvent(ctx, config.ProductAddedTopic, productAddedProducer, newProductAdded)
		produceOrderingEvent(ctx, config.ProductRemovedTopic, productRemovedProducer, newProductRemoved)
		produceOrderingEvent(ctx, config.CartViewedTopic, cartViewedProducer, newCartViewed)
		produceOrderingEvent(ctx, config.CheckoutStartedTopic, checkoutStartedProducer, newCheckoutStarted)
		produceOrderingEvent(ctx, config.CheckoutStepViewedTopic, checkoutStepViewedProducer, newCheckoutStepViewed)
		produceOrderingEvent(ctx, config.CheckoutStepCompletedTopic, checkoutStepCompletedProducer, newCheckoutStepCompleted)
		produceOrderingEvent(ctx, config.PaymentInfoEnteredTopic, paymentInfoEnteredProducer, newPaymentInfoEntered)
		produceOrderingEvent(ctx, config.OrderUpdatedTopic, orderUpdatedProducer, newOrderUpdated)
		produceOrderingEvent(ctx, config.OrderCompletedTopic, orderCompletedProducer, newOrderCompleted)
		produceOrderingEvent(ctx, config.OrderRefundedTopic, orderRefundedProducer, newOrderRefunded)
		produceOrderingEvent(ctx, config.OrderCancelledTopic, orderCancelledProducer, newOrderCancelled)

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
		time.Sleep(time.Second)
	}
}

// produceOrderingEvent is a helper function to produce an ordering event.
func produceOrderingEvent[M proto.Message](ctx context.Context, topic string, producer *produce.Producer[M], generator func(string) M) {
	if topic == "" {
		return
	}
	id := newID()
	if err := producer.ProduceProtobufMessage(ctx, id, generator(id)); err != nil {
		if errors.Is(err, context.Canceled) {
			// Context cancelled, likely shutting down
			return
		}
		slog.Error("error producing event", "topic", topic, "error", err)
	} else {
		slog.Info("produced event", "topic", topic, "id", id)
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
	products := generateDummyProducts(gofakeit.Number(1, 5))
	return &demov1.ProductListFiltered{
		Id:       id,
		ListId:   gofakeit.UUID(),
		Filters:  filters,
		Sorts:    sorts,
		Products: products,
	}
}

// generateDummyProducts generates a slice of dummy products.
func generateDummyProducts(count int) []*demov1.Product {
	products := make([]*demov1.Product, count)
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
	return products
}

// --- Ordering Event Generators ---

func newProductClicked(id string) *demov1.ProductClicked {
	return &demov1.ProductClicked{
		ProductId: gofakeit.UUID(),
		Sku:       gofakeit.Word(),
		Category:  gofakeit.Word(),
		Name:      gofakeit.ProductName(),
		Brand:     gofakeit.Company(),
		Variant:   gofakeit.Color(),
		Price:     float64(gofakeit.Price(10, 200)),
		Quantity:  1,
		Coupon:    gofakeit.Word(),
		Position:  int32(gofakeit.Number(1, 10)),
		Url:       gofakeit.URL(),
		ImageUrl:  gofakeit.URL(),
	}
}

func newProductViewed(id string) *demov1.ProductViewed {
	return &demov1.ProductViewed{
		ProductId: gofakeit.UUID(),
		Sku:       gofakeit.Word(),
		Category:  gofakeit.Word(),
		Name:      gofakeit.ProductName(),
		Brand:     gofakeit.Company(),
		Variant:   gofakeit.Color(),
		Price:     float64(gofakeit.Price(10, 200)),
		Quantity:  1,
		Coupon:    gofakeit.Word(),
		Currency:  "USD",
		Position:  int32(gofakeit.Number(1, 10)),
		Url:       gofakeit.URL(),
		ImageUrl:  gofakeit.URL(),
	}
}

func newProductAdded(id string) *demov1.ProductAdded {
	return &demov1.ProductAdded{
		CartId:    gofakeit.UUID(),
		ProductId: gofakeit.UUID(),
		Sku:       gofakeit.Word(),
		Category:  gofakeit.Word(),
		Name:      gofakeit.ProductName(),
		Brand:     gofakeit.Company(),
		Variant:   gofakeit.Color(),
		Price:     float64(gofakeit.Price(10, 200)),
		Quantity:  int32(gofakeit.Number(1, 3)),
		Coupon:    gofakeit.Word(),
		Position:  int32(gofakeit.Number(1, 10)),
		Url:       gofakeit.URL(),
		ImageUrl:  gofakeit.URL(),
	}
}

func newProductRemoved(id string) *demov1.ProductRemoved {
	return &demov1.ProductRemoved{
		CartId:    gofakeit.UUID(),
		ProductId: gofakeit.UUID(),
		Sku:       gofakeit.Word(),
		Category:  gofakeit.Word(),
		Name:      gofakeit.ProductName(),
		Brand:     gofakeit.Company(),
		Variant:   gofakeit.Color(),
		Price:     float64(gofakeit.Price(10, 200)),
		Quantity:  int32(gofakeit.Number(1, 3)),
		Coupon:    gofakeit.Word(),
		Position:  int32(gofakeit.Number(1, 10)),
		Url:       gofakeit.URL(),
		ImageUrl:  gofakeit.URL(),
	}
}

func newCartViewed(id string) *demov1.CartViewed {
	return &demov1.CartViewed{
		CartId:   gofakeit.UUID(),
		Products: generateDummyProducts(gofakeit.Number(1, 5)),
	}
}

func newCheckoutStarted(id string) *demov1.CheckoutStarted {
	products := generateDummyProducts(gofakeit.Number(1, 5))
	total := 0.0
	for _, p := range products {
		total += p.Price
	}
	return &demov1.CheckoutStarted{
		OrderId:     gofakeit.UUID(),
		Affiliation: gofakeit.Company(),
		Value:       total * 1.1, // Include potential tax/shipping
		Revenue:     total,
		Shipping:    total * 0.05,
		Tax:         total * 0.05,
		Discount:    0,
		Coupon:      gofakeit.Word(),
		Currency:    "USD",
		Products:    products,
	}
}

func newCheckoutStepViewed(id string) *demov1.CheckoutStepViewed {
	return &demov1.CheckoutStepViewed{
		CheckoutId:     gofakeit.UUID(),
		Step:           int32(gofakeit.Number(1, 3)),
		ShippingMethod: gofakeit.RandomString([]string{"UPS", "FedEx", "USPS"}),
		PaymentMethod:  gofakeit.CreditCardType(),
	}
}

func newCheckoutStepCompleted(id string) *demov1.CheckoutStepCompleted {
	return &demov1.CheckoutStepCompleted{
		CheckoutId:     gofakeit.UUID(),
		Step:           int32(gofakeit.Number(1, 3)),
		ShippingMethod: gofakeit.RandomString([]string{"UPS", "FedEx", "USPS"}),
		PaymentMethod:  gofakeit.CreditCardType(),
	}
}

func newPaymentInfoEntered(id string) *demov1.PaymentInfoEntered {
	products := generateDummyProducts(gofakeit.Number(1, 5))
	total := 0.0
	for _, p := range products {
		total += p.Price
	}
	return &demov1.PaymentInfoEntered{
		CheckoutId: gofakeit.UUID(),
		OrderId:    gofakeit.UUID(),
		Total:      total * 1.1,
		Revenue:    total,
		Shipping:   total * 0.05,
		Tax:        total * 0.05,
		Discount:   0,
		Coupon:     gofakeit.Word(),
		Currency:   "USD",
		Products:   products,
	}
}

func newOrderUpdated(id string) *demov1.OrderUpdated {
	products := generateDummyProducts(gofakeit.Number(1, 5))
	total := 0.0
	for _, p := range products {
		total += p.Price
	}
	return &demov1.OrderUpdated{
		OrderId:     gofakeit.UUID(),
		Affiliation: gofakeit.Company(),
		Total:       total * 1.1,
		Revenue:     total,
		Shipping:    total * 0.05,
		Tax:         total * 0.05,
		Discount:    0,
		Coupon:      gofakeit.Word(),
		Currency:    "USD",
		Products:    products,
	}
}

func newOrderCompleted(id string) *demov1.OrderCompleted {
	products := generateDummyProducts(gofakeit.Number(1, 5))
	subtotal := 0.0
	for _, p := range products {
		subtotal += p.Price
	}
	shipping := subtotal * 0.05
	tax := subtotal * 0.05
	discount := 0.0
	total := subtotal + shipping + tax - discount
	return &demov1.OrderCompleted{
		CheckoutId:  gofakeit.UUID(),
		OrderId:     gofakeit.UUID(),
		Affiliation: gofakeit.Company(),
		Total:       total,
		Subtotal:    subtotal,
		Revenue:     subtotal - discount, // Revenue often excludes tax/shipping
		Shipping:    shipping,
		Tax:         tax,
		Discount:    discount,
		Coupon:      gofakeit.Word(),
		Currency:    "USD",
		Products:    products,
	}
}

func newOrderRefunded(id string) *demov1.OrderRefunded {
	products := generateDummyProducts(gofakeit.Number(1, 2)) // Usually partial refund
	totalRefund := 0.0
	for _, p := range products {
		totalRefund += p.Price
	}
	return &demov1.OrderRefunded{
		OrderId:  gofakeit.UUID(),
		Total:    totalRefund,
		Currency: "USD",
		Products: products,
	}
}

func newOrderCancelled(id string) *demov1.OrderCancelled {
	products := generateDummyProducts(gofakeit.Number(1, 5))
	total := 0.0
	for _, p := range products {
		total += p.Price
	}
	return &demov1.OrderCancelled{
		OrderId:     gofakeit.UUID(),
		Affiliation: gofakeit.Company(),
		Total:       total * 1.1,
		Revenue:     total, // Or 0 depending on cancellation policy
		Shipping:    total * 0.05,
		Tax:         total * 0.05,
		Discount:    0,
		Coupon:      gofakeit.Word(),
		Currency:    "USD",
		Products:    products,
	}
}
