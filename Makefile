.PHONY: help build test test-integration lint \
	web-backend-build web-backend-test web-backend-test-integration web-backend-lint \
	bot-test frontend-test

help:
	@echo "make build              build every service that exists"
	@echo "make test               unit test every service that exists"
	@echo "make test-integration   integration test every service that exists"
	@echo "make lint               lint every service that exists"

build: web-backend-build

test: web-backend-test bot-test frontend-test

test-integration: web-backend-test-integration

lint: web-backend-lint

web-backend-build:
	$(MAKE) -C web/backend build

web-backend-test:
	$(MAKE) -C web/backend test

web-backend-test-integration:
	$(MAKE) -C web/backend test-integration-full

web-backend-lint:
	$(MAKE) -C web/backend lint

bot-test:
	@if [ -d bot ]; then $(MAKE) -C bot test; else echo "skip: bot/ not present yet"; fi

frontend-test:
	@if [ -d web/frontend ]; then $(MAKE) -C web/frontend test; else echo "skip: web/frontend/ not present yet"; fi
