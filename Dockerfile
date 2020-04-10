# builder
FROM golang:latest AS builder
RUN mkdir /app
ADD . /app/
WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main .

# certificates
FROM alpine:latest as certs
RUN apk --no-cache add ca-certificates

# runner
FROM scratch
#WORKDIR /app

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/main /main
# Workaround, cf: https://github.com/markbates/pkger/issues/86
COPY --from=builder /app/templates /templates/

ENTRYPOINT ["/main"]
