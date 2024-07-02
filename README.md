# Bufstream Demo

## Requirements

- Git
- [Docker](https://docs.docker.com/engine/install/)
- [Buf CLI](https://buf.build/docs/installation)
- [Go 1.22+](https://go.dev/doc/install) _[optional, for local development]_

## Setup steps

Before starting, perform the following steps to prepare your environment to run 
the demo.

### Getting started

1. Clone this repo:

   ```shell
   git clone https://github.com/bufbuild/bufstream-demo.git
   cd bufstream-demo
   ```

1. Start downloading Docker images: _[optional]_

   ```shell
   docker-compose pull
   ```

## Bufstream Demo

The Bufstream demo application simulates a small Connect service that manages 
email addresses. Email addresses can be updated, but must be verified to ensure 
they're valid. (Imagine the service enqueuing up change events that a separate 
process consumes to fire off verification emails.)

The demo publishes email updated events to a Bufstream topic when changes are 
made, while a separate goroutine consumes the same topic and "verifies" the 
change. The demo app uses an off-the-shelf Kafka client to interact with 
Bufstream.  

### Run the demo app

1. Boot up the Bufstream, AKHQ, and demo apps:

   ```shell
   docker-compose up --build
   ```

1. In another terminal window, send a request to update an email address:

   ```shell
   curl -X POST -H "content-type:application/json" \
       -d '{ "uuid": "53b61d8e-420b-4588-bbc4-bb2325099922", "new_address": "demo@buf.build" }' \
       localhost:8888/bufbuild.bufstream_demo.v1.EmailService/UpdateEmail
   ```

1. The app should log both the publish of the event on the producer side and
   shortly after the verification of the email address. You can see the email
   is marked as verified:

   ```shell
   curl -X POST -H "content-type:application/json" \
       -d '{ "uuid": "53b61d8e-420b-4588-bbc4-bb2325099922" }' \
       localhost:8888/bufbuild.bufstream_demo.v1.EmailService/GetEmail
   ```
   
1. Toggle off the verifier to enqueue update events, before re-enabling to watch 
   the consumer work through the backlog.

   ```shell
   curl -X POST -H "content-type:application/json" \
       -d '{ "enabled": false }' \
       localhost:8888/bufbuild.bufstream_demo.v1.EmailService/ToggleVerifier
   ```

### Introspect via AKHQ

1. Navigate to http://localhost:8080 to view the AKHQ dashboard.

1. The topic used by the demo should be visible in AKHQ, along with the 
   verifier's consumer group, and the record(s) published should be present. 
   They will be encoded in the binary wire format and may be somewhat illegible.

## Data Enforcement

As is, the service does not apply any semantic enforcement on the data passing 
through it; UUIDs and addresses can be invalid or malformed. To ensure the 
quality of data flowing through Bufstream, data enforcement policies that can be
configured to require data to conform to a pre-determined schema, pass semantic 
validation, and even redact sensitive information. (Note that this same 
validation could and should be performed by the service itself, but for 
demonstration purposes, we will only perform validation on the data flowing 
through Bufstream.)

### Setting up the BSR

1. Login to your BSR web app using your credentials.

1. Create a user token on the BSR web app to log into the Buf CLI. *Note*: Save this token somewhere you can easily retrieve it, we'll be using it in another step later.

   ```shell
   buf registry login <BSR_HOST> --username <BSR_USERNAME>
   ```

1. Uncomment the `GOPRIVATE` environment variable and `.netrc` lines in the [`Dockerfile`](./Dockerfile), replacing `BSR_HOST`, `BSR_USER`, and `BSR_TOKEN`. These values can be found in your local `~/.netrc`.

   ```dockerfile
   ENV GOPRIVATE "<BSR_HOST>/gen/go"
   RUN echo "machine <BSR_HOST> login <BSR_USER> password <BSR_TOKEN>" > ~/.netrc
   ```

### Creating the CSR instance

1. Navigate to the admin panel in the BSR to create a CSR instance under Confluent integration. Name it something unique to you.

1. Uncomment and update `BSR_HOST`, `CSR_INSTANCE`, `BSR_USER`, and `BSR_TOKEN` in the `demo` service of the [`docker-compose.yaml`](docker-compose.yaml) with the information you created earlier.

   ```yaml
   command: [
     "--address", "0.0.0.0:8888",
     "--bootstrap", "bufstream:9092",
     "--csr-url", "https://<BSR_HOST>/integrations/confluent/<CSR_INSTANCE>",
     "--csr-user", "<BSR_USER>",
     "--csr-pass", "<BSR_TOKEN>",
   ]
   ```
   
1. Uncomment and update [`configs/akhq.yaml`](configs/akhq.yaml) the `schema-registry` section with the same information:

   ```yaml
    schema-registry:
      type: "confluent"
      url: "https://<BSR_HOST>/integrations/confluent/<CSR_INSTANCE>"
      basic-auth-username: "<BSR_USER>"
      basic-auth-password: "<BSR_TOKEN>"
   ```

1. Uncomment and update [`configs/bufstream.yaml`](configs/bufstream.yaml) the `data_enforcement` section with the same information:

   ```yaml
     data_enforcement:
       schema_registries:
         - name: csr
           confluent:
             url: "http://<BSR_HOST>/integrations/confluent/<CSR_INSTANCE>"
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

### Creating the BSR Repository and CSR Subject/Schema

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
        name: "<BSR_HOST>/<BSR_USER>/<BSR_REPO>"
    ```

1. Ensure dependencies are locked:

   ```shell
   buf dep update
   ```

1. Generate the app code:

   ```shell
   buf generate
   ```

1. Create and push the repository up to the BSR (and associated CSR instance):

   ```shell
   buf push --create --create-visibility public
   ```

1. Stop & restart Docker services to pick up the changes:

    ```shell
    # ctrl+c to gracefully terminate current services
    docker-compose up --build
    ```

### Semantic validation

1. Attempt to update an email with invalid data. While the service permits such 
   malformed data, Bufstream will reject produced data that does not match the 
   schema and semantic validation rules.

   ```shell
   curl -X POST -H "content-type:application/json" \
       -d '{ "uuid": "foo", "new_address": "bar" }' \
       localhost:8888/bufbuild.bufstream_demo.v1.EmailService/UpdateEmail
   ```

1. Because the event was rejected, the verifier does not verify the address 
   change.

   ```shell
   curl -X POST -H "content-type:application/json" \
       -d '{ "uuid": "foo" }' \
       localhost:8888/bufbuild.bufstream_demo.v1.EmailService/GetEmail
   ```

### Redaction

1. Successfully update an email. Because the data is valid, production of the 
   event and later verification will succeed.

   ```shell
   curl -X POST -H "content-type:application/json" \
       -d '{ "uuid": "53b61d8e-420b-4588-bbc4-bb2325099922", "new_address": "demo@buf.build" }' \
       localhost:8888/bufbuild.bufstream_demo.v1.EmailService/UpdateEmail
   ```
   
2. Because we marked the `old_address` field of the event as redacted, a second 
   update of the same email address will not print the `old_address` in the logs.

   ```shell
   curl -X POST -H "content-type:application/json" \
       -d '{ "uuid": "53b61d8e-420b-4588-bbc4-bb2325099922", "new_address": "bufstream@buf.build" }' \
       localhost:8888/bufbuild.bufstream_demo.v1.EmailService/UpdateEmail
   ```

3. Because AKHQ also receives the redacted form of the messages, `old_address` 
   will not be present in the records there either.   

## Cleanup

After completing the demo, you can stop Docker and remove credentials from your
machine with the following commands:

```
buf registry logout <BSR_HOST>
docker-compose down --rmi all
```