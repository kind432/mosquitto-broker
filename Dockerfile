FROM golang:1.20 AS builder

WORKDIR /mqtt_broker

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd;

FROM debian:bookworm-slim AS final

RUN apt-get update && apt-get install -y --no-install-recommends \
    mosquitto \
    mosquitto-clients \
    ca-certificates \
    procps \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /mqtt_broker

COPY --from=builder /mqtt_broker/app .
COPY configs ./configs
COPY mosquitto-data ./mosquitto-data

RUN chmod -R 777 /mqtt_broker

ENV MOSQUITTO_PORT=1882

EXPOSE 1882 8000

CMD ["./app"]