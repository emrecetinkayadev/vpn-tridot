.PHONY: build test run migrate

build:
	go build -o bin/api ./cmd/api

test:
	go test ./...

run:
	APP_ENV=local go run ./cmd/api

migrate:
	GOOSE_DRIVER?=postgres
	GOOSE_DBSTRING?=
	@if [ -z "$(GOOSE_DBSTRING)" ]; then \
		echo "GOOSE_DBSTRING is required" >&2; \
		exit 1; \
	fi
	goose -dir ./internal/storage/postgres/migrations up
