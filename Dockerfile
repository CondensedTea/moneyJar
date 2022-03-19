FROM golang:1.17 as builder

WORKDIR /build

COPY Makefile ./
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY pkg/ ./pkg/

RUN make build

FROM scratch

COPY --from=builder /build/bin/moneyJar .

ENTRYPOINT ["./moneyJar"]