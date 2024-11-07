FROM golang:1.20 AS builder

ENV GOPATH=/

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN go build -o app ./main.go

CMD ["./app" ]