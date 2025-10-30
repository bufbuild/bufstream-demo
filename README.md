# Bufstream demo

Bufstream is the Kafka-compatible message queue built for the data lakehouse era. Building on top of off-the-shelf technologies such as S3 and Postgres instead of expensive machines with large attached disks, Bufstream is 8x less expensive to operate than Apache Kafka. But that's not why we built it: Bufstream brings [schema-driven development](https://buf.build/blog/kafka-schema-driven-development) and schema governance to streaming data, solving the data quality problems that plague Kafka.

Paired with the [Buf Schema Registry](../bsr/index.md), Bufstream's **broker-side schema awareness** solves longstanding problems with data quality and schema governance while enabling new capabilities like **semantic validation** and **direct-to-Iceberg topic storage**.

This repository contains code used in [Bufstream's quickstart](https://buf.build/docs/bufstream/quickstart).
Head over to the quickstart to walk through this repository and get started!

## Curious to see more?

To learn more about Bufstream, check out the
[introduction](https://buf.build/docs/bufstream/) or
[join us in the Buf Slack](https://buf.build/links/slack)!
