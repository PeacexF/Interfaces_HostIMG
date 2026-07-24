.PHONY: help build run test test-integration test-integration-full integration-db-up integration-db-down migrate-up migrate-down sqlc lint vulncheck

MIGRATE ?= migrate
DB_DSN ?= postgres://postgres:postgres@localhost:5432/interfaces_web?sslmode=disable
TEST_DB_DSN ?= postgres://postgres:postgres@localhost:5433/interfaces_web_test?sslmode=disable

INTEGRATION_PG_CONTAINER ?= interfaces-web-test-pg
INTEGRATION_PG_IMAGE ?= postgres:16-alpine
INTEGRATION_PG_PORT ?= 5433
INTEGRATION_PG_DB ?= interfaces_web_test

help:
	@echo "  make build                 build the server binary"
	@echo "  make run                   build + run"
	@echo "  make test                  unit tests"
	@echo "  make migrate-up            apply migrations (DB_DSN)"
	@echo "  make migrate-down          rollback last migration"
	@echo "  make sqlc                  regenerate sqlc code"
	@echo "  make integration-db-up     start + migrate a disposable Postgres"
	@echo "  make integration-db-down   stop it"
	@echo "  make test-integration      run integration tests (needs integration-db-up)"
	@echo "  make test-integration-full up + test + down in one shot"
	@echo "  make lint                  gofmt + go vet"
	@echo "  make vulncheck             govulncheck"

build:
	go build -o bin/server ./cmd/server

run: build
	./bin/server

test:
	go test ./...

migrate-up:
	$(MIGRATE) -path db/migrations -database "$(DB_DSN)" up

migrate-down:
	$(MIGRATE) -path db/migrations -database "$(DB_DSN)" down 1

sqlc:
	sqlc generate

integration-db-up:
	@docker rm -f $(INTEGRATION_PG_CONTAINER) >/dev/null 2>&1 || true
	docker run -d --name $(INTEGRATION_PG_CONTAINER) \
		-p $(INTEGRATION_PG_PORT):5432 \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=$(INTEGRATION_PG_DB) \
		$(INTEGRATION_PG_IMAGE)
	@for i in $$(seq 1 30); do \
		if docker exec $(INTEGRATION_PG_CONTAINER) pg_isready -U postgres -d $(INTEGRATION_PG_DB) >/dev/null 2>&1; then \
			$(MIGRATE) -path db/migrations -database "$(TEST_DB_DSN)" up; \
			exit 0; \
		fi; \
		sleep 1; \
	done; \
	echo "postgres did not become ready"; exit 1

integration-db-down:
	@docker rm -f $(INTEGRATION_PG_CONTAINER) >/dev/null 2>&1 || true

test-integration:
	TEST_DATABASE_DSN="$(TEST_DB_DSN)" go test -tags=integration ./... -run Integration -v

test-integration-full: integration-db-up
	@status=0; \
	$(MAKE) test-integration || status=$$?; \
	$(MAKE) integration-db-down; \
	exit $$status

lint:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then echo "$$unformatted"; exit 1; fi
	go vet ./...

vulncheck:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
