# Bufstream Demo

## Requirements

- Git
- [Docker](https://docs.docker.com/engine/install/)
- [Buf CLI](https://buf.build/docs/installation)
- [Go 1.22+](https://go.dev/doc/install) _[optional, for local development]_
- Data enforcement section: Private [Enterprise Buf Schema Registry (BSR) server](https://buf.build/docs/bsr/private/overview)
  and admin-level access

## Getting started

Before starting, perform the following steps to prepare your environment to run
the demo.

1. Clone this repo:

   ```shellsession
   $ git clone https://github.com/bufbuild/bufstream-demo.git
   $ cd bufstream-demo
   ```

1. Start downloading Docker images: _[optional]_

   ```shellsession
   $ docker-compose pull
   ```

## Bufstream demo

The Bufstream demo application simulates a small Connect service that manages
email addresses. Email addresses can be updated, but must be verified to ensure
they're valid. (Imagine the service enqueuing change events that a separate
process consumes to send verification emails.)

The demo publishes "email updated" events to a Bufstream topic when changes are
made, while a separate goroutine consumes the same topic and "verifies" the
change. The demo app uses an off-the-shelf Kafka client to interact with
Bufstream.

### Run the demo app

1. Boot up the Bufstream, AKHQ, and demo apps:

   ```shellsession
   $ docker-compose up --build
   ```

1. In another terminal window, send a request to update an email address:

   ```shellsession
   $ curl -X POST -H "content-type:application/json" \
       -d '{ "uuid": "53b61d8e-420b-4588-bbc4-bb2325099922", "new_address": "demo@buf.build" }' \
       localhost:8888/bufbuild.bufstream_demo.v1.EmailService/UpdateEmail
   ```

1. The app should log both the publishing of the event on the producer side and
   shortly after, the verification of the email address. You can see the email
   is marked as verified:

   ```shellsession
   $ curl -X POST -H "content-type:application/json" \
       -d '{ "uuid": "53b61d8e-420b-4588-bbc4-bb2325099922" }' \
       localhost:8888/bufbuild.bufstream_demo.v1.EmailService/GetEmail
   ```

1. Toggle off the verifier to enqueue update events, before re-enabling to watch
   the consumer work through the backlog.

   ```shellsession
   $ curl -X POST -H "content-type:application/json" \
       -d '{ "enabled": false }' \
       localhost:8888/bufbuild.bufstream_demo.v1.EmailService/ToggleVerifier
   ```

### Introspect via AKHQ

Navigate to http://localhost:8080 to view the AKHQ dashboard. The topic used by
the demo should be visible in AKHQ, along with the verifier's consumer group,
and the record(s) published should be present. They're encoded in the binary
wire format and may be somewhat illegible.

## Data enforcement demo

As is, the service doesn't apply any semantic enforcement on the data passing
through it—UUIDs and addresses can be invalid or malformed. To ensure the
quality of data flowing through Bufstream, you can configure data enforcement
policies that require data to conform to a pre-determined schema, pass semantic
validation, and even redact sensitive information. (Note that this same
validation could and should be performed by the service itself, but for
demonstration purposes, we will only perform validation on the data flowing
through Bufstream.)

**Note:** This part of the demo uses the BSR Confluent integration to
demonstrate schema enforcement and validation, but any Confluent Schema Registry
implementation can be used with Bufstream.

### Set up the BSR

1. Sign in to your BSR server using your credentials.

1. Create a user token in the BSR to log into the Buf CLI, then log in. Save this token somewhere you can easily retrieve it—you'll be using it in another step later.

   ```shellsession
   $ buf registry login <BSR_SERVER> --username <BSR_USERNAME>
   ```

1. Uncomment the `GOPRIVATE` environment variable and `.netrc` lines in the [`Dockerfile`](./Dockerfile), replacing `BSR_SERVER`, `BSR_USER`, and `BSR_TOKEN`. These values can be found in your local `~/.netrc` file.

   ```dockerfile
   ENV GOPRIVATE "<BSR_SERVER>/gen/go"
   RUN echo "machine <BSR_SERVER> login <BSR_USER> password <BSR_TOKEN>" > ~/.netrc
   ```

### Create the CSR instance

1. Navigate to the admin panel in the BSR to create a CSR instance under **Confluent integration**. Name it something unique to you.

1. Uncomment and update `BSR_SERVER`, `CSR_INSTANCE`, `BSR_USER`, and `BSR_TOKEN` in the `demo` service of the [`docker-compose.yaml`](./docker-compose.yaml) file with the information you've created.

   ```yaml
   command: [
       "--address", "0.0.0.0:8888",
       "--bootstrap", "bufstream:9092",
       "--csr-url", "https://<BSR_SERVER>/integrations/confluent/<CSR_INSTANCE>",
       "--csr-user", "<BSR_USER>",
       "--csr-pass", "<BSR_TOKEN>",
   ]
   ```

1. Uncomment and update the `schema-registry` section of the [`configs/akhq.yaml`](configs/akhq.yaml) file with the same information:

   ```yaml
   schema-registry:
     type: "confluent"
     url: "https://<BSR_SERVER>/integrations/confluent/<CSR_INSTANCE>"
     basic-auth-username: "<BSR_USER>"
     basic-auth-password: "<BSR_TOKEN>"
   ```

1. Uncomment and update the `data_enforcement` section of the [`configs/bufstream.yaml`](configs/bufstream.yaml) file with the same information:

   ```yaml
   data_enforcement:
     schema_registries:
       - name: csr
         confluent:
           url: "http://<BSR_SERVER>/integrations/confluent/<CSR_INSTANCE>"
           instance_name: "<CSR_INSTANCE>"
           basic_auth:
             username:
               string: "<BSR_USER>"
             password:
               string: "<BSR_TOKEN>"
     produce:
       - topics: { all: true }
         schema_registry: csr
         values:
           coerce: true
           on_parse_error: REJECT_BATCH
           validation:
             on_error: REJECT_BATCH
     fetch:
       - topics: { all: true }
         schema_registry: csr
         values:
           on_parse_error: FILTER_RECORD
           redaction:
             debug_redact: true
   ```

### Create the BSR repository and CSR subject/schema

1. Update [`proto/bufbuild/bufstream_demo/v1/events.proto`](proto/bufbuild/bufstream_demo/v1/events.proto) with the CSR instance name you created previously:

   ```protobuf
     option (buf.confluent.v1.subject) = {
       instance_name: "<CSR_INSTANCE>"
       name: "email-updated-value"
     };
   ```

1. Uncomment and update [`buf.yaml`](buf.yaml) with a unique `name` for the repository:

   ```yaml
   modules:
     - path: proto
       name: "<BSR_SERVER>/<BSR_USER>/<BSR_REPO>"
   ```

1. Ensure that your dependencies are locked:

   ```shellsession
   $ buf dep update
   ```

1. Generate the app code:

   ```shellsession
   $ buf generate
   ```

1. Create and push the repository up to the BSR (and associated CSR instance):

   ```shellsession
   $ buf push --create --create-visibility public
   ```

1. Stop and restart Docker services to pick up the changes (Ctrl+c to gracefully terminate current services):

   ```shellsession
   $ docker-compose up --build
   ```

### Semantic validation

This part of the demo shows how integration with a CSR enables Bufstream to reject
data that doesn't match the schema and semantic validation rules, before it
enters the data stream.

1. Attempt to update an email with invalid data. Though the service permits such
   malformed data, Bufstream rejects it.

   ```shellsession
   $ curl -X POST -H "content-type:application/json" \
       -d '{ "uuid": "foo", "new_address": "bar" }' \
       localhost:8888/bufbuild.bufstream_demo.v1.EmailService/UpdateEmail
   ```

1. Because the event was rejected, the verifier does not verify the address
   change.

   ```shellsession
   $ curl -X POST -H "content-type:application/json" \
       -d '{ "uuid": "foo" }' \
       localhost:8888/bufbuild.bufstream_demo.v1.EmailService/GetEmail
   ```

### Redaction

This part of the demo shows how having the `fetch.redaction.debug_redact` field set
to `true` prevents PII (the old email address in this case) from entering the data
stream. That setting interacts with the field setting in the `events.proto`
file:

```protobuf {10}
syntax = "proto3";

package bufbuild.bufstream_demo.v1;

import "buf/confluent/v1/extensions.proto";
import "buf/validate/validate.proto";

message EmailUpdated {
  string uuid = 1 [(buf.validate.field).string.uuid = true];
  string old_address = 2 [
    (buf.validate.field).string.email = true,
    (buf.validate.field).ignore = IGNORE_EMPTY,
    debug_redact = true
  ];
  string new_address = 3 [(buf.validate.field).string.email = true];

  option (buf.confluent.v1.subject) = {
    instance_name: "<CSR_INSTANCE>"
    name: "email-updated"
  };
}
```

Successfully update an email. Because the data is valid, production of the
event and later verification succeed.

```shellsession
$ curl -X POST -H "content-type:application/json" \
      -d '{ "uuid": "53b61d8e-420b-4588-bbc4-bb2325099922", "new_address": "demo@buf.build" }' \
      localhost:8888/bufbuild.bufstream_demo.v1.EmailService/UpdateEmail
```

Because the `old_address` field of the event is marked as redacted, a second
update of the same email address doesn't print the `old_address` in the logs.

```shellsession
$ curl -X POST -H "content-type:application/json" \
      -d '{ "uuid": "53b61d8e-420b-4588-bbc4-bb2325099922", "new_address": "bufstream@buf.build" }' \
      localhost:8888/bufbuild.bufstream_demo.v1.EmailService/UpdateEmail
```

AKHQ also receives the redacted form of the messages, so `old_address`
isn't present in the records there either.

## Cleanup

After completing the demo, you can stop Docker and remove credentials from your
machine with the following commands:

```
buf registry logout <BSR_SERVER>
docker-compose down --rmi all
```
