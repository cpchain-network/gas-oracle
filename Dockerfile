FROM golang:1.24-alpine3.21 as builder

RUN apk add --no-cache make ca-certificates gcc musl-dev linux-headers git jq bash

COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum

WORKDIR /app

RUN go mod download

# build gas-oracle with the shared go.mod & go.sum files
COPY . /app/gas-oracle

WORKDIR /app/gas-oracle

RUN make gas-oracle

FROM alpine:3.18

COPY --from=builder /app/gas-oracle/gas-oracle /usr/local/bin
COPY --from=builder /app/gas-oracle/gas-oracle.yaml /app/gas-oracle/gas-oracle.yaml
COPY --from=builder /app/gas-oracle/migrations /app/gas-oracle/migrations

ENV INDEXER_MIGRATIONS_DIR="/app/gas-oracle/migrations"
WORKDIR /app/gas-oracle
