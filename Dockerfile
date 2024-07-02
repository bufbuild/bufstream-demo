FROM golang:1.22 AS builder

#ENV GOPRIVATE "<BSR_HOST>/gen/go"
#RUN echo "machine <BSR_HOST> login <BSR_USER> password <BSR_TOKEN>" > ~/.netrc
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./
RUN go build -trimpath -o ./bufstream-demo .

EXPOSE 8888
ENTRYPOINT ["/app/bufstream-demo"]