package csr

import (
	"errors"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type ClientConfig struct {
	URL      string
	Username string
	Password string
}

func NewCSRClient(clientConfig ClientConfig) (schemaregistry.Client, error) {
	return schemaregistry.NewClient(newCSRConfig(clientConfig))
}

func NewSingleTypeProtobufSerializer[M proto.Message]() serde.Serializer {
	return singleTypeProtobufSerializer[M]{}
}

func NewSingleTypeProtobufDeserializer[M proto.Message]() serde.Deserializer {
	return singleTypeProtobufDeserializer[M]{}
}

func NewCSRProtobufSerializer(csrClient schemaregistry.Client) (serde.Serializer, error) {
	return protobuf.NewSerializer(csrClient, serde.ValueSerde, protobuf.NewSerializerConfig())
}

func NewCSRProtobufDeserializer(csrClient schemaregistry.Client) (serde.Deserializer, error) {
	deserializer, err := protobuf.NewDeserializer(csrClient, serde.ValueSerde, protobuf.NewDeserializerConfig())
	if err != nil {
		return nil, err
	}
	deserializer.ProtoRegistry = protoregistry.GlobalTypes
	return deserializer, nil
}

func newCSRConfig(clientConfig ClientConfig) *schemaregistry.Config {
	if clientConfig.Username != "" && clientConfig.Password != "" {
		return schemaregistry.NewConfigWithBasicAuthentication(
			clientConfig.URL,
			clientConfig.Username,
			clientConfig.Password,
		)
	}
	return schemaregistry.NewConfig(clientConfig.URL)
}

type singleTypeProtobufSerializer[M proto.Message] struct{}

func (singleTypeProtobufSerializer[M]) ConfigureSerializer(
	schemaregistry.Client,
	serde.Type,
	*serde.SerializerConfig,
) error {
	// TODO: why not error like deserializer?
	return nil
}

func (singleTypeProtobufSerializer[M]) Serialize(_ string, value interface{}) ([]byte, error) {
	message, ok := value.(M)
	if !ok {
		return nil, fmt.Errorf("unknown message type: %T", value)
	}
	return proto.Marshal(message)
}

func (singleTypeProtobufSerializer[M]) Close() error {
	return nil
}

type singleTypeProtobufDeserializer[M proto.Message] struct{}

func (singleTypeProtobufDeserializer[M]) ConfigureDeserializer(
	schemaregistry.Client,
	serde.Type,
	*serde.DeserializerConfig,
) error {
	return errors.New("unimplemented")
}

func (singleTypeProtobufDeserializer[M]) Deserialize(_ string, payload []byte) (interface{}, error) {
	var message M
	var ok bool
	message, ok = message.ProtoReflect().Type().New().Interface().(M)
	if !ok {
		return nil, fmt.Errorf("did not get message type %T from ProtoReflect", message)
	}
	if err := proto.Unmarshal(payload, message); err != nil {
		return nil, err
	}
	return message, nil
}

func (singleTypeProtobufDeserializer[M]) DeserializeInto(_ string, payload []byte, value interface{}) error {
	message, ok := value.(M)
	if !ok {
		return fmt.Errorf("unknown message type: %T", value)
	}
	return proto.Unmarshal(payload, message)
}

func (singleTypeProtobufDeserializer[M]) Close() error {
	return nil
}
