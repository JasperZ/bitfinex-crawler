
############################
# STEP 1 build executable binary
############################
FROM golang:1.12.9-alpine3.10 AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git
# Create appuser.
RUN adduser -D -g '' appuser
WORKDIR $GOPATH/src/github.com/jasperz/bitfinex-crawler
COPY . .
# Fetch dependencies.
# Using go get.
RUN go get -d -v
# Using go mod.
# RUN go mod download
# RUN go mod verify
# Build the binary.
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/bitfinex-crawler

############################
# STEP 2 build a small image
############################
FROM scratch
# Import the user and group files from the builder.
COPY --from=builder /etc/passwd /etc/passwd
# Copy our static executable.
COPY --from=builder /go/bin/bitfinex-crawler /go/bin/bitfinex-crawler
# Use an unprivileged user.
USER appuser
# Run the hello binary.
ENTRYPOINT ["/go/bin/bitfinex-crawler"]