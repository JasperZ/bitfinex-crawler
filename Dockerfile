FROM golang:1.12.9-alpine3.10

ADD . /go/src/github.com/jasperz/bitfinex-crawler

RUN apk add git
RUN go get github.com/jasperz/bitfinex-crawler
RUN go install github.com/jasperz/bitfinex-crawler

ENTRYPOINT /go/bin/bitfinex-crawler