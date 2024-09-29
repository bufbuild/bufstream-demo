package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	demov1 "github.com/bufbuild/bufstream-demo/gen/bufstream/demo/v1"
	"github.com/bufbuild/bufstream-demo/pkg/csr"
	"github.com/bufbuild/bufstream-demo/pkg/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/spf13/pflag"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := run(ctx); err != nil {
		slog.Error("program error", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg := ParseConfig()

	client, err := kafka.NewKafkaClient(
		kafka.Config{
			BootstrapServers: cfg.bootstrapServers,
			Group:            cfg.groupID,
			Topic:            cfg.topic,
			ClientID:         "bufstream-demo",
		},
	)
	if err != nil {
		return err
	}
	defer client.Close()

	var serializer serde.Serializer
	var deserializer serde.Deserializer
	if cfg.csrURL != "" {
		csrClient, err := csr.NewCSRClient(cfg.csrURL, cfg.csrUser, cfg.csrPass)
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

	producer := NewProducer(client, cfg.topic, serializer)
	consumer := NewConsumer[*demov1.EmailUpdated](
		client,
		deserializer,
		cfg.topic,
		WithMessageHandler(handleEmailUpdated),
	)
	for {
		n, err := producer.Run(ctx)
		if err != nil {
			return fmt.Errorf("produce error: %w", err)
		}

		if err := consumer.Consume(ctx, n); err != nil && !errors.Is(err, context.Canceled) {
			return fmt.Errorf("consume error: %w", err)
		}
		time.Sleep(time.Second)
	}
}

type Config struct {
	bootstrapServers []string
	groupID          string
	topic            string
	address          string
	csrURL           string
	csrUser          string
	csrPass          string
}

func ParseConfig() (cfg Config) {
	cfg = Config{
		bootstrapServers: []string{"bufstream:9092"},
		topic:            "email-updated",
		groupID:          "email-verifier",
		address:          "0.0.0.0:8888",
		csrURL:           "",
		csrUser:          "",
		csrPass:          "",
	}

	pflag.StringArrayVarP(&cfg.bootstrapServers, "bootstrap", "b", cfg.bootstrapServers, "Bufstream bootstrap server(s)")
	pflag.StringVarP(&cfg.topic, "topic", "t", cfg.topic, "Email updates topic name")
	pflag.StringVarP(&cfg.groupID, "group", "g", cfg.groupID, "Consumer consumer group ID")
	pflag.StringVarP(&cfg.address, "address", "a", cfg.address, "Email service listener address")
	pflag.StringVarP(&cfg.csrURL, "csr-url", "c", cfg.csrURL, "CSR URL")
	pflag.StringVarP(&cfg.csrUser, "csr-user", "u", cfg.csrUser, "CSR username")
	pflag.StringVarP(&cfg.csrPass, "csr-pass", "p", cfg.csrPass, "CSR password/token")
	pflag.Parse()

	return cfg
}

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
