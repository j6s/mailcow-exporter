FROM golang:1.15 as builder

COPY ./ /build
RUN cd /build \
    && go build -o /mailcow-exporter /build/main.go \
    && rm -Rf /build

FROM alpine:3.13

RUN apk add --no-cache \
        openssl \
        ca-certificates \
        libc6-compat
COPY --from=builder /mailcow-exporter /usr/local/bin/mailcow-exporter

ENTRYPOINT mailcow-exporter
