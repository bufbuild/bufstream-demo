# Bufstream Demo

## What is Bufstream?

[Bufstream](https://buf.build/product/bufstream) is a drop-in replacement for Apache Kafka that's
10x cheaper to operate. While Bufstream works well for any Kafka workload, it excels when paired
with Protobuf. By integrating with the Buf Schema Registry, Bufstream can enforce data quality and
governance policies without relying on opt-in client configuration.

In this demonstration, you'll run a Bufstream agent locally, use the
[`franz-go`](https://github.com/twmb/franz-go) library to publish and consume messages, and explore
Bufstream's schema enforcement features.

## Requirements

- Git
- [Docker](https://docs.docker.com/engine/install)
- [Buf CLI](https://buf.build/docs/installation)
- [Go 1.22+](https://go.dev/doc/install) _[optional, for local development]_

## Getting started

Before starting, perform the following steps to prepare your environment to run the demo.

1. Clone this repo:

   ```shellsession
   $ git clone https://github.com/bufbuild/bufstream-demo.git
   $ cd ./bufstream-demo
   ```

## Try out Bufstream

The Bufstream demo application simulates two small applications that produce and consume
[`EmailUpdated`](proto/bufstream/demo/v1/demo.proto) event messages. The demo
[producer](cmd/bufstream-demo-produce/main.go) publishes these events to a Bufstream topic, followed
by a separate [consumer](cmd/bufstream-demo-consume/main.go) that fetches from the same topic and
"verifies" the change. The demo app uses the off-the-shelf Kafka client
[`franz-go`](https://github.com/twmb/franz-go) to interact with Bufstream.

The demo attempts to publish and consume three payloads:

- A semantically valid, correctly formatted version of the `EmailUpdated` message.
- A correctly formatted, but semantically invalid version of the message.
- A malformed message.

### Replace Apache Kafka

1. Boot up the Bufstream and demo apps:

   ```shellsession
   $ docker compose up --build
   ```

1. The app logs both the publishing of the events on the producer side and shortly after, the
   consumption of these events. The consumer deserializes the first two messages correctly, while
   the final message fails with a deserialization error.

At this point, you've used Bufstream as a drop-in replacement for Apache Kafka.

### Data quality enforcement

Thus far, Bufstream hasn't applied any quality enforcement on the data passing through it &mdash;
addresses can be invalid or malformed. To ensure the quality of data flowing through Bufstream, you
can configure policies that require data to conform to a pre-determined schema, pass semantic
validation, and even redact sensitive information on the fly.

We'll demonstrate this functionality using the Buf Schema Registry, but Bufstream also works with
any registry that implements Confluent's REST API.

1. Uncomment the `data_enforcement` block in [`config/bufstream.yaml`](config/bufstream.yaml).

1. Uncomment the `--csr` command options under the `demo` service in
   [`docker-compose.yaml`](docker-compose.yaml).

1. To pick up the configuration changes, terminate the Docker apps via `ctrl+c` and restart them
   with `docker compose up`.

1. The app again logs both the attempted publishing and consumption of the three events. The first
   event will successfully reach the consumer. But with Bufstream's data quality enforcement
   enabled, the second and third messages are rejected.

1. For the message that reaches the consumer, notice the empty `old_address` in the logs. This field
   has been redacted by Bufstream as it's labeled with the `debug_redact` option.

By configuring Bufstream with a few data quality and governance policies, you've ensured that
consumers receive only well-formed, semantically valid data.

### Cleanup

After completing the demo, you can stop Docker and remove credentials from your machine with the
following commands:

```
docker compose down --rmi all
```

## Curious to see more?

To learn more about Bufstream, check out the
[launch blog post](https://buf.build/blog/bufstream-kafka-lower-cost), dig into the
[benchmark and cost analysis](https://buf.build/docs/bufstream/cost), or
[join us in the Buf Slack](https://buf.build/links/slack)!
