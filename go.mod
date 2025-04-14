module github.com/bufbuild/bufstream-demo

go 1.23.0

toolchain go1.24.0

require (
	buf.build/gen/go/bufbuild/confluent/protocolbuffers/go v1.36.5-20240926213411-65369e65bbcd.1
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.5-20250219170025-d39267d9df8f.1
	github.com/brianvoe/gofakeit/v7 v7.2.1
	github.com/confluentinc/confluent-kafka-go/v2 v2.8.0
	github.com/google/uuid v1.6.0
	github.com/spf13/pflag v1.0.6
	github.com/twmb/franz-go v1.18.1
	github.com/twmb/franz-go/pkg/kadm v1.15.0
	google.golang.org/protobuf v1.36.5
)

require (
	github.com/bufbuild/protocompile v0.14.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/jhump/protoreflect v1.17.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.9.0 // indirect
	golang.org/x/crypto v0.35.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	google.golang.org/genproto v0.0.0-20250224174004-546df14abb99 // indirect
)
