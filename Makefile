LOCAL_BIN:=$(CURDIR)/bin

VERSION:=0.1

.PHONY: build
build:
	CGO_ENABLED=0 go build -o bin/moneyJar ./cmd/moneyJar/main.go

.PHONY: run
run:
	CGO_ENABLED=0 go run ./cmd/moneyJar/main.go --loglevel=debug

.PHONY: lint
lint:
	 golangci-lint run --config=.golangci.yaml ./...

.PHONY: docker.up
local.up:
	docker-compose up -d

.PHONY: docker.down
local.down:
	docker-compose down

.PHONY: docker.build
local.build:
	docker-compose build
