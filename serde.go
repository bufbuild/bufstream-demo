package main

import (
	"errors"
	"fmt"

	demov1 "github.com/bufbuild/bufstream-demo/internal/gen/bufstream/demo/v1"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func NewSerde(csrURL, username, password string) (serde.Serializer, serde.Deserializer, error) {
	if csrURL == "" {
		protoSerde := ProtoSerde[*demov1.EmailUpdated]{}
		return protoSerde, protoSerde, nil
	}
	srClient, err := NewSchemaRegistryClient(csrURL, username, password)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize schema registry client: %w", err)
	}
	serializer, err := NewSerializer(srClient)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize serializer: %w", err)
	}
	deserializer, err := NewDeserializer(srClient)
	if err != nil {
		serializer.Close()
		return nil, nil, fmt.Errorf("failed to initialize deserializer: %w", err)
	}
	return serializer, deserializer, nil
}

func NewSchemaRegistryClient(csrURL, username, password string) (schemaregistry.Client, error) {
	cfg := schemaregistry.NewConfigWithAuthentication(
		csrURL, username, password)
	return schemaregistry.NewClient(cfg)
}

func NewSerializer(client schemaregistry.Client) (serde.Serializer, error) {
	return protobuf.NewSerializer(client, serde.ValueSerde, protobuf.NewSerializerConfig())
}

func NewDeserializer(client schemaregistry.Client) (serde.Deserializer, error) {
	des, err := protobuf.NewDeserializer(client, serde.ValueSerde, protobuf.NewDeserializerConfig())
	if err != nil {
		return nil, err
	}
	des.ProtoRegistry = protoregistry.GlobalTypes
	return des, nil
}

type ProtoSerde[M proto.Message] struct{}

func (p ProtoSerde[M]) ConfigureSerializer(_ schemaregistry.Client, _ serde.Type, _ *serde.SerializerConfig) error {
	return nil
}

func (p ProtoSerde[M]) Serialize(_ string, data interface{}) ([]byte, error) {
	if msg, ok := data.(M); ok {
		return proto.Marshal(msg)
	}
	return nil, fmt.Errorf("unknown data type: %T", data)
}

func (p ProtoSerde[M]) ConfigureDeserializer(schemaregistry.Client, serde.Type, *serde.DeserializerConfig) error {
	return errors.New("unimplemented")
}

func (p ProtoSerde[M]) Deserialize(_ string, payload []byte) (interface{}, error) {
	var msg M
	msg = msg.ProtoReflect().Type().New().Interface().(M)
	return msg, proto.Unmarshal(payload, msg)
}

func (p ProtoSerde[M]) DeserializeInto(_ string, payload []byte, data interface{}) error {
	msg, ok := data.(M)
	if !ok {
		return fmt.Errorf("unknown data type: %T", data)
	}
	return proto.Unmarshal(payload, msg)
}

func (p ProtoSerde[M]) Close() error {
	return nil
}
