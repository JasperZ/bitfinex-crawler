FROM golang:1.12.7

ADD . /go/src/github.com/jasperz/bitfinex-crawler

RUN go install github.com/jasperz/bitfinex-crawler

ENTRYPOINT /go/bin/bitfinex-crawler