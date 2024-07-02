package storage

import (
	"context"

	bufstream_demov1 "github.com/bufbuild/bufstream-demo/gen/bufbuild/bufstream_demo/v1"
)

type Interface interface {
	GetEmail(ctx context.Context, uuid string) (*bufstream_demov1.UserEmail, error)
	UpdateEmail(ctx context.Context, uuid, address string) (previous string, err error)
	VerifyEmail(ctx context.Context, uuid, address string) error
}
