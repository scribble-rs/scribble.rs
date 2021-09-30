FROM golang:latest AS builder
RUN mkdir /app
ADD . /app/
WORKDIR /app

RUN make build

FROM alpine:latest as certs
RUN apk --no-cache add ca-certificates

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/scribblers /scribblers

ENTRYPOINT ["/scribblers"]
