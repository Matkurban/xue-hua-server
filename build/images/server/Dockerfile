# Multi-service image build for xue-hua-server (adapted from openimsdk/open-im-server).
FROM golang:1.24-alpine AS builder
ARG CMD_PATH
ARG BINARY_NAME
ARG RELEASE=false
WORKDIR /server
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN set -eu; \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" \
      -o "/out/${BINARY_NAME}" "./${CMD_PATH}"

FROM alpine:3.23
ARG BINARY_NAME
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /server
COPY --from=builder /out/${BINARY_NAME} /server/bin/${BINARY_NAME}
COPY config /server/config
ENV PATH="/server/bin:${PATH}"
