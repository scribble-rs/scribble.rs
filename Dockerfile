FROM golang:1.12.5-stretch
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go mod download
RUN go build -o main .
CMD ["/app/main", "--portHTTP=80"]