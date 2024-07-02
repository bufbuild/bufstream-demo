package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"

	"github.com/spf13/pflag"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg := ParseConfig()

	client := Must(NewKafkaClient(cfg.bootstrapServers, cfg.groupID, cfg.topic))
	defer client.Close()

	serializer, deserializer, err := NewSerde(cfg.csrURL, cfg.csrUser, cfg.csrPass)
	Must[any](nil, err)
	defer serializer.Close()
	defer deserializer.Close()

	producer := NewProducer(client, cfg.topic, serializer)
	if err := producer.Run(ctx); err != nil {
		slog.Error("producer exited with error", "error", err)
		os.Exit(1)
	}

	consumer := NewConsumer(client, cfg.topic, deserializer)
	slog.Info("consumer started", "topic", cfg.topic)
	if err := consumer.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("consumer exited with error", "error", err)
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

func Must[T any](v T, err error) T {
	if err != nil {
		slog.Error("initialization error", "error", err)
		os.Exit(1)
	}
	return v
}
