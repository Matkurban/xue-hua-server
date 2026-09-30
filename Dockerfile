FROM golang:1.27.1-alpine AS builder
WORKDIR /server
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN set -eu; mkdir -p /server/bin; \
    for source in cmd/server-*/main.go cmd/server-rpc/server-*/main.go tools/check-component/main.go tools/check-free-memory/main.go tools/seq/main.go; do \
      directory="${source%/main.go}"; binary="${directory##*/}"; \
      CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X server/v3/version.Version=${VERSION}" \
        -o "/server/bin/${binary}" "./${directory}"; \
    done

FROM alpine:3.23
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /server
COPY --from=builder /server/bin ./bin
COPY --from=builder /server/config ./config
COPY --from=builder /server/LICENSE ./LICENSE
ENV PATH="/server/bin:${PATH}"
CMD ["/server/bin/server-api", "-c", "/server/config"]
