// Package csr implements helper functionality around the Confluent Schema Registry.
package csr

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/sr"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Config is all the configuration needed to connect to a CSR Instance.
//
// Note that the schemaregistry package has its own NewConfig.* functions, which we call.
// However, we're bringing this down to exactly what we need for this demo.
type Config struct {
	// The URL of the CSR instance.
	//
	// The absence of this field says to not connect to the CSR.
	URL string
	// The username to use for authentication, if any.
	Username string
	// The password to use for authentication, if any.
	Password string
}

// Serde is a serializer/deserializer.
type Serde interface {
	// Encode encodes value into bytes.
	Encode(value any) (bytes []byte, err error)
	// DecodeNew decodes bytes into value.
	DecodeNew(bytes []byte) (value any, err error)
}

// NewSerde creates a new serializer/deserializer.
func NewSerde[M proto.Message](ctx context.Context, config Config, topic string) (Serde, error) {
	if config.URL != "" {
		return newCSRSerde[M](ctx, config, topic)
	}
	return basicSerde[M]{}, nil
}

// newCSRSerde creates a new Serde backed by the given CSR configured in the config.
func newCSRSerde[M proto.Message](ctx context.Context, config Config, topic string) (*sr.Serde, error) {
	client, err := sr.NewClient(
		sr.URLs(config.URL),
		sr.BasicAuth(config.Username, config.Password),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create schema registry client: %w", err)
	}
	schema, err := client.SchemaByVersion(ctx, topic+"-value", -1)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema for %q: %w", topic, err)
	}

	serde := sr.NewSerde()
	var msg M
	var index []int
	var desc protoreflect.Descriptor = msg.ProtoReflect().Descriptor()
	for {
		index = append(index, desc.Index())
		desc = desc.Parent()
		if _, ok := desc.(protoreflect.FileDescriptor); ok {
			break
		}
	}

	basic := basicSerde[M]{}
	serde.Register(schema.ID, msg,
		sr.Index(index...),
		sr.GenerateFn(func() any { return msg.ProtoReflect().New().Interface() }),
		sr.EncodeFn(basic.Encode),
		sr.DecodeFn(basic.Decode),
	)
	return serde, nil
}

// basicSerde implements Serde for a [proto.Message].
type basicSerde[M proto.Message] struct{}

// Encode encodes val.
func (b basicSerde[M]) Encode(val any) ([]byte, error) {
	msg, ok := val.(M)
	if !ok {
		return nil, fmt.Errorf("expected a %T, got %T", msg, val)
	}
	return proto.Marshal(msg)
}

// Decode decodes bytes into val.
func (b basicSerde[M]) Decode(bytes []byte, val any) error {
	msg, ok := val.(M)
	if !ok {
		return fmt.Errorf("expected a %T, got %T", msg, val)
	}
	return proto.Unmarshal(bytes, msg)
}

// DecodeNew decodes bytes into a new M.
func (b basicSerde[M]) DecodeNew(bytes []byte) (any, error) {
	var msg M
	msg, _ = msg.ProtoReflect().New().Interface().(M)
	return msg, b.Decode(bytes, msg)
}
