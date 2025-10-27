module github.com/bufbuild/bufstream-demo

go 1.25

require (
	buf.build/gen/go/bufbuild/confluent/protocolbuffers/go v1.36.10-20240926213411-65369e65bbcd.1
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.10-20250912141014-52f32327d4b0.1
	github.com/brianvoe/gofakeit/v7 v7.8.0
	github.com/google/uuid v1.6.0
	github.com/spf13/pflag v1.0.10
	github.com/twmb/franz-go v1.20.1
	github.com/twmb/franz-go/pkg/kadm v1.17.1
	github.com/twmb/franz-go/pkg/sr v1.5.0
	google.golang.org/protobuf v1.36.10
)

require (
	github.com/klauspost/compress v1.18.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.12.0 // indirect
	golang.org/x/crypto v0.43.0 // indirect
)
