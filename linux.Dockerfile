#
# Builder for Golang
#
# We explicitly use a certain major version of go, to make sure we don't build
# with a newer verison than we are using for CI tests, as we don't directly
# test the produced binary but from source code directly.
FROM docker.io/golang:1.23.3 AS builder

WORKDIR /app

# This causes caching of the downloaded go modules and makes repeated local
# builds much faster. We must not copy the code first though, as a change in
# the code causes a redownload.
COPY go.mod go.sum ./
RUN go mod download -x

# Import that this comes after mod download, as it breaks caching.
ARG VERSION="dev"

# Copy actual codebase, since we only have the go.mod and go.sum so far.
COPY . /app/
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags "-w -s -X 'github.com/scribble-rs/scribble.rs/internal/version.Version=${VERSION}'" -tags timetzdata -o ./scribblers ./cmd/scribblers

#
# Runner
#
FROM scratch

COPY --from=builder /app/scribblers /scribblers
# The scratch image doesn't contain any certificates, therefore we use
# the builders certificate, so that we can send HTTP requests.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/scribblers"]
# Random uid to avoid having root privileges. Linux doesn't care that there's no user for it.
USER 248:248
