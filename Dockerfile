FROM golang:1.23-alpine3.20 as builder

WORKDIR /app

COPY go.mod go.sum /app/
RUN go mod download

COPY . /app
RUN go build -ldflags "-s -w" -trimpath -buildvcs=false -o /go/bin/bufstream-demo .

FROM scratch

COPY --from=builder /go/bin/bufstream-demo /bufstream-demo
ENTRYPOINT ["/bufstream-demo"]
