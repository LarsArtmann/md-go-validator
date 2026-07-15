FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOEXPERIMENT=jsonv2 go build -o /md-go-validator ./cmd/md-go-validator

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=builder /md-go-validator /usr/local/bin/md-go-validator

ENTRYPOINT ["md-go-validator"]
