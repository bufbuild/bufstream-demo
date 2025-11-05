# bufstream-demo

[Bufstream](https://buf.build/product/bufstream) is a fully self-hosted drop-in replacement for Apache Kafka® that writes data to S3-compatible object storage. It's 100% compatible with the Kafka protocol, including support for exactly-once semantics (EOS) and transactions. Bufstream is 8x cheaper to operate, and a single cluster can elastically scale to hundreds of GB/s of throughput. It's the universal Kafka replacement for the modern age.

Additionally, for teams sending Protobuf messages across their Kafka topics, Bufstream is a perfect partner. Bufstream can enforce data quality and governance requirements on the broker with [Protovalidate](https://github.com/bufbuild/protovalidate). Bufstream can directly persist records as [Apache Iceberg™](https://iceberg.apache.org/) tables, reducing time-to-insight in popular data lakehouse products such as Snowflake or ClickHouse.

This repository contains code used in [Bufstream's quickstart](https://buf.build/docs/bufstream/quickstart), which steps through and explains Bufstream's capabilities.

If you're here for a quick test drive or to demonstrate Bufstream:

1. Clone this repository and navigate to its directory.
2. Keep reading to learn how to use the included `make` targets.

## Using the `make` targets

### Semantic validation with [Go](https://go.dev/) installed

1. Use `make bufstream-run` to download and run Bufstream's single binary.
2. In a second terminal, run `make produce-run` to produce sample e-commerce shopping cart messages.
3. In a third terminal, run `make consume-run` to start consuming messages. About 1% contain semantically invalid messages and cause errors.
4. Stop the producer. Run `make use-reject-mode`. Restart the producer: it logs errors when it tries to produce invalid messages. The consumer soon stops logging errors about invalid carts. 
5. Stop the producer. Run `make use-dlq-mode`. Restart the producer: there are no more errors. All invalid messages are sent to a DLQ topic.
6. In a fourth terminal, run `make consume-dlq-run`. It reads the `orders.dlq` topic and shows that the original message can be reconstructed and examined.
7. Stop all processes before continuing to Iceberg.

### Semantic validation without [Go](https://go.dev/) installed

If you don't have Go installed, you can still run this demonstration via a Docker Compose project.

1. Use `make docker-compose-run` to start the Compose project. The producer immediately begins producing sample e-commerce shopping cart messages. About 1% of the messages are semantically invalid and cause the consumer to log errors.   
2. Open a second terminal and run `make docker-compose-use-reject-mode`. Back in the first terminal, invalid messages are now rejected: the producer logs errors and the consumer stops receiving any invalid messages. 
3. In the second terminal, run `make docker-compose-use-dlq-mode`. The producer stops receiving errors, and the DLQ consumer begins logging invalid messages sent to the `orders.dlq` topic.
4. Stop the Compose project and use `make docker-compose-clean` before continuing to Iceberg.

### Iceberg integration

The Iceberg demo uses the Docker Compose project defined in [./iceberg/docker-compose.yaml](./iceberg/docker-compose.yaml) to provide services such as an Iceberg catalog and Spark.

1. Run `make iceberg-run` to start the Iceberg project. The Spark image is a large download, and there are multiple services to start. When you see `create-orders-topic-1 exited with code 0`, continue.
2. In a different terminal, run `make iceberg-produce` to create sample data. Once you've produced about 1,000 records, stop the process.
3. Run `make iceberg-table` to manually run the Bufstream task that updates Iceberg catalogs. 
4. Open http://localhost:8888/notebooks/notebooks/bufstream-quickstart.ipynb, click within the SELECT query's cell, and use shift-return or the ▶︎ icon to build a revenue report based on the `orders` topic.
5. This example is durable: the Compose project can be stopped and started without losing data. To remove all data and images, stop the Compose project and run `make iceberg-clean`.

## Curious to see more?

To learn more about Bufstream, check out the
[launch blog post](https://buf.build/blog/bufstream-kafka-lower-cost), dig into the
[benchmark and cost analysis](https://buf.build/docs/bufstream/cost), or
[join us in the Buf Slack](https://buf.build/links/slack)!
