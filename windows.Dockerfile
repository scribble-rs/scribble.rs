#
# Builder for Golang
#
# We explicitly use a certain major version of go, to make sure we don't build
# with a newer verison than we are using for CI tests, as we don't directly
# test the produced binary but from source code directly.
FROM docker.io/golang:1.20.7-nanoserver-ltsc2022 AS builder
WORKDIR /app

# This causes caching of the downloaded go modules and makes repeated local
# builds much faster. We must not copy the code first though, as a change in
# the code causes a redownload.
COPY go.mod go.sum ./
RUN go mod download -x

# Copy actual codebase, since we only have the go.mod and go.sum so far.
COPY . /app/
ENV CGO_ENABLED=0
RUN go build -tags timetzdata -o ./scribblers ./cmd/scribblers

#
# Runner
#
FROM mcr.microsoft.com/windows/nanoserver:ltsc2022

COPY --from=builder /app/scribblers /scribblers

# Required so go knows which timezone to use by default, if none is
# explicitly defined when using the `time` package.
ENV TZ=Europe/Berlin

ENTRYPOINT ["/scribblers"]
