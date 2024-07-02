FROM golang:1.22 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./
RUN go build -trimpath -o ./bufstream-demo .
ENTRYPOINT ["/app/bufstream-demo"]