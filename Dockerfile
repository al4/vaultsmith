# Build stage
FROM golang:1.10.3-alpine3.8 as builder

ENV GOPATH /go
ENV CGO_ENABLED 0

RUN apk update && apk add curl git
RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && chmod +x /usr/local/bin/dep
WORKDIR /go/src/github.com/starlingbank/vaultsmith
COPY Gopkg.toml Gopkg.lock ./

RUN dep ensure -vendor-only
COPY . .
RUN go build -a --installsuffix cgo --ldflags="-s"

# Run tests
FROM builder as tester
RUN go test ./...

# Production image stage
FROM alpine:3.8

RUN apk --update upgrade

RUN rm -rf /var/cache/apk/*

COPY --from=builder /go/src/github.com/starlingbank/vaultsmith/vaultsmith .

ENTRYPOINT ["./vaultsmith"]
CMD ["-h"]
