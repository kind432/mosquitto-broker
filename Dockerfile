FROM golang:1.20 AS builder

WORKDIR /mqtt_broker

COPY . .

RUN go mod download

RUN go build -o app ./main.go

FROM debian:bookworm AS final

RUN apt-get update && apt-get install -y \
    mosquitto \
    mosquitto-clients \
    procps && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /mqtt_broker

COPY --from=builder /mqtt_broker /mqtt_broker

RUN mkdir -p /mqtt_broker/internal/mosquitto && \
    touch /mqtt_broker/internal/mosquitto/mosquitto.log && \
    chmod 666 /mqtt_broker/internal/mosquitto/mosquitto.log

ENV MOSQUITTO_PORT=1882

CMD ["/mqtt_broker/app"]
