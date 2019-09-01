
############################
# STEP 1 build executable binary
############################
FROM golang:1.12.9-alpine3.10 AS builder
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/github.com/jasperz/bitfinex-crawler
COPY . .
RUN go get -d -v
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/bitfinex-crawler

############################
# STEP 2 build a small image
############################
FROM alpine:3.10
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/bitfinex-crawler /go/bin/bitfinex-crawler
RUN /go/bin/bitfinex-crawler
ENTRYPOINT ["/go/bin/bitfinex-crawler"]