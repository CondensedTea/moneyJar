FROM golang:1.17-alpine as builder

WORKDIR /build

RUN apk update && apk add --no-cache ca-certificates && apk add --no-cache make
RUN update-ca-certificates

COPY Makefile ./
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY pkg/ ./pkg/
COPY config.yaml ./
COPY messages.yaml ./

RUN make build

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/bin/moneyJar .
COPY --from=builder /build/config.yaml .
COPY --from=builder /build/messages.yaml .

ENTRYPOINT ["./moneyJar"]