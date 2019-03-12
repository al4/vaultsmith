# Build stage
FROM golang:1.11.5-alpine3.8 as builder

ENV CGO_ENABLED 0

RUN apk update && apk add curl git tar

RUN curl -fsSlL -o - https://github.com/alecthomas/gometalinter/releases/download/v2.0.12/gometalinter-2.0.12-linux-amd64.tar.gz | \
    tar -xvz --strip-components=1 -C /usr/local/bin

WORKDIR /src/vaultsmith

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -a --installsuffix cgo --ldflags="-s"

# Run tests
FROM builder as tester
RUN go test ./...
# gometalinter, with a long deadline. Use shorter times (~60s) locally.
#RUN gometalinter --deadline=240s --vendor --enable-gc --tests -I '^talebearer' /src/talebearer
#RUN gometalinter --deadline=240s --vendor --enable-gc --tests /src/talebearer/internal

# Production image stage
FROM alpine:3.8

RUN apk --no-cache --update upgrade

COPY --from=builder /src/vaultsmith/vaultsmith /usr/local/bin

ENTRYPOINT ["/usr/local/bin/vaultsmith"]
CMD ["-h"]
