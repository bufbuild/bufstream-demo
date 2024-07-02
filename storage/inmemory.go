package storage

import (
	"context"
	"fmt"
	"sync"

	"connectrpc.com/connect"
	bufstream_demov1 "github.com/bufbuild/bufstream-demo/gen/bufbuild/bufstream_demo/v1"
	"google.golang.org/protobuf/proto"
)

type InMemory struct {
	mu     sync.RWMutex
	emails map[string]*bufstream_demov1.UserEmail
}

func NewInMemory() *InMemory {
	return &InMemory{
		emails: make(map[string]*bufstream_demov1.UserEmail),
	}
}

func (s *InMemory) GetEmail(_ context.Context, uuid string) (*bufstream_demov1.UserEmail, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.emails[uuid]
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound,
			fmt.Errorf("no email found for UUID %q", uuid))
	}
	return proto.Clone(val).(*bufstream_demov1.UserEmail), nil
}

func (s *InMemory) UpdateEmail(_ context.Context, uuid, address string) (previous string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.emails[uuid]
	previous = val.GetEmailAddress()
	if !ok {
		val = &bufstream_demov1.UserEmail{
			Uuid: uuid,
		}
		s.emails[uuid] = val
	}
	val.EmailAddress = address
	val.Verified = false
	return previous, nil
}

func (s *InMemory) VerifyEmail(_ context.Context, uuid, address string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.emails[uuid]
	switch {
	case !ok:
		return connect.NewError(connect.CodeNotFound,
			fmt.Errorf("no email found for UUID %q", uuid))
	case val.GetEmailAddress() != address:
		return connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("registered email address does not match provided address %q", address))
	default:
		val.Verified = true
		return nil
	}
}
