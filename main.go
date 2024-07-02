package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/bufbuild/bufstream-demo/storage"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
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

	var store storage.Interface = storage.NewInMemory()
	store = storage.NewEventEmitter(store, cfg.topic, serializer, client)

	verifier := NewVerifier(store, cfg.topic, deserializer, client)
	service := NewEmailService(store, client)

	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error { return service.Run(ctx, cfg.address) })
	grp.Go(func() error { return verifier.Run(ctx) })

	if err := grp.Wait(); err != nil {
		slog.Error("runtime error", "error", err)
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
	pflag.StringVarP(&cfg.groupID, "group", "g", cfg.groupID, "Verifier consumer group ID")
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
