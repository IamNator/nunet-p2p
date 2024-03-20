FROM golang:1.21-alpine as builder

WORKDIR /go/src/github.com/iamnator/nunet-p2p

COPY . .

RUN go get -d -v ./...

RUN go build -o /go/bin/app

# Path: Dockerfile
FROM alpine:3.7

COPY --from=builder /go/bin/app /app

CMD ["/app"]