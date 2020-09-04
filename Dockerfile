FROM golang:1.15

COPY ./ /build
RUN cd /build \
    && go build -o /mailcow-exporter /build/src/*.go \
    && rm -Rf /build

CMD /mailcow-exporter --host=$MAILCOW_HOST --api-key=$MAILCOW_API_KEY
