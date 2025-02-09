# bufstream-demo

[Bufstream](https://buf.build/product/bufstream) is a fully self-hosted drop-in replacement for Apache Kafka® that writes data to S3-compatible object storage. It’s 100% compatible with the Kafka protocol, including support for exactly-once semantics (EOS) and transactions. Bufstream is 8x cheaper to operate, and a single cluster can elastically scale to hundreds of GB/s of throughput. It's the universal Kafka replacement for the modern age.

Additionally, for teams sending Protobuf messages across their Kafka topics, Bufstream is a perfect partner. Bufstream can enforce data quality and governance requirements on the broker with [Protovalidate](https://github.com/bufbuild/protovalidate). Bufstream can directly persist records as [Apache Iceberg™](https://iceberg.apache.org/) tables, reducing time-to-insight in popular data lakehouse products such as Snowflake or ClickHouse.

This repository contains code used in [Bufstream's quickstart](https://buf.build/docs/bufstream/quickstart).
Head over to the quickstart to walk through this repository and get started!

## Curious to see more?

To learn more about Bufstream, check out the
[launch blog post](https://buf.build/blog/bufstream-kafka-lower-cost), dig into the
[benchmark and cost analysis](https://buf.build/docs/bufstream/cost), or
[join us in the Buf Slack](https://buf.build/links/slack)!
