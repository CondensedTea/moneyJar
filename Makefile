LOCAL_BIN:=$(CURDIR)/bin

DSN :=

.PHONY: build
build:
	CGO_ENABLED=0 go build -o bin/moneyJar ./cmd/moneyJar/main.go

.PHONY: run
run:
	CGO_ENABLED=0 go run ./cmd/moneyJar/main.go

.PHONY: local.up
local.up:
	docker-compose up -d

.PHONY: local.down
local.down:
	docker-compose down

.PHONY: local.migrate
local.migrate:
	goose up
