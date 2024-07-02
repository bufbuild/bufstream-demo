package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/connect"
	v1 "github.com/bufbuild/bufstream-demo/gen/bufbuild/bufstream_demo/v1"
	"github.com/bufbuild/bufstream-demo/gen/bufbuild/bufstream_demo/v1/bufstream_demov1connect"
	"github.com/bufbuild/bufstream-demo/storage"
	"github.com/twmb/franz-go/pkg/kgo"
)

type EmailService struct {
	store  storage.Interface
	client *kgo.Client
}

func NewEmailService(store storage.Interface, client *kgo.Client) *EmailService {
	return &EmailService{
		store:  store,
		client: client,
	}
}

func (e *EmailService) Run(ctx context.Context, address string) error {
	mux := http.NewServeMux()
	mux.Handle(bufstream_demov1connect.NewEmailServiceHandler(e))
	srv := &http.Server{
		Addr:    address,
		Handler: mux,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	slog.InfoContext(ctx, "email service started", "listener", address)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (e *EmailService) GetEmail(ctx context.Context, c *connect.Request[v1.GetEmailRequest]) (*connect.Response[v1.GetEmailResponse], error) {
	val, err := e.store.GetEmail(ctx, c.Msg.GetUuid())
	if err != nil {
		return nil, err
	}
	res := &v1.GetEmailResponse{UserEmail: val}
	return connect.NewResponse(res), nil
}

func (e *EmailService) UpdateEmail(ctx context.Context, c *connect.Request[v1.UpdateEmailRequest]) (*connect.Response[v1.UpdateEmailResponse], error) {
	_, err := e.store.UpdateEmail(ctx, c.Msg.GetUuid(), c.Msg.GetNewAddress())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&v1.UpdateEmailResponse{}), nil
}

func (e *EmailService) ToggleVerifier(ctx context.Context, c *connect.Request[v1.ToggleVerifierRequest]) (*connect.Response[v1.ToggleVerifierResponse], error) {
	if c.Msg.GetEnabled() {
		slog.InfoContext(ctx, "enabled verifier")
		e.client.ResumeFetchTopics(e.client.PauseFetchTopics()...)
	} else {
		slog.InfoContext(ctx, "disabled verifier")
		e.client.PauseFetchTopics(e.client.GetConsumeTopics()...)
	}
	return connect.NewResponse(&v1.ToggleVerifierResponse{}), nil
}
