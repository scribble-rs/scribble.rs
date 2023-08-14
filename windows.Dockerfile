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
# We have to use a debian image, because with alpine we get an error
#   Error loading shared library libresolv.so.2: No such file or directory (needed by /service)
#   Error relocating /flows_service: __res_search: symbol not found
# The following tips for alpine based images didn't solve the problem
#   RUN apk add libaio libnsl libc6-compat
#   RUN ln -s /lib64/* /lib
#   RUN ln -s /lib/libc.so.6 /usr/lib/libresolv.so.2
#
# This seems to be related to the fact, that simply passing `CGO_ENABLED=0`
# will not necessarily produce a static build. Whether this is enough to
# produce a static build, depends on both the target os and the libraries used.
# For example the standard net package seems to require C libraries.
#
# People are currently working on providing a simpler way of producing
# static binaries:
# https://github.com/golang/go/issues/26492
FROM mcr.microsoft.com/windows/nanoserver:ltsc2022

COPY --from=builder /app/scribblers /scribblers

# # Certificiates aren't installed by default, but are required for https request.
# RUN apt update && apt install ca-certificates -y

# Required so go knows which timezone to use by default, if none is
# explicitly defined when using the `time` package.
ENV TZ=Europe/Berlin

ENTRYPOINT ["/scribblers"]
