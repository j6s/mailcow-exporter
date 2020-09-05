FROM golang:1.15

COPY ./ /build
RUN cd /build \
    && go build -o /mailcow-exporter /build/main.go \
    && rm -Rf /build

ENTRYPOINT /mailcow-exporter
