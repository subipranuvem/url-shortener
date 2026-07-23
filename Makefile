.PHONY: help deps run build routes docker-up docker-down docker-reset migrate test compile-docs validate-docs docs-up docs-down validate-krakend seed

KRAKEND_VERSION := 2.13.8

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_0-9-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

deps: ## Download and tidy dependencies
	go mod tidy

run: ## Run the API locally (requires .env)
	go run ./cmd/api

build: ## Build the binary
	go build -ldflags="-s -w" -o shortener ./cmd/api

routes: ## Print all registered API endpoints
	@go run ./cmd/routes

docker-up: ## Build and start all containers
	KRAKEND_VERSION=$(KRAKEND_VERSION) docker compose up --build -d

docker-down: ## Stop and remove all containers
	KRAKEND_VERSION=$(KRAKEND_VERSION) docker compose down

docker-reset: ## Tear down all containers and volumes, then rebuild (fixes stale ScyllaDB state)
	KRAKEND_VERSION=$(KRAKEND_VERSION) docker compose down -v
	KRAKEND_VERSION=$(KRAKEND_VERSION) docker compose up --build -d

migrate: ## Run CQL migrations against local ScyllaDB
	go run ./cmd/migrate -dir migrations

test: ## Run all tests
	go test ./...

compile-docs: ## Compile OpenAPI docs into a single file (requires Docker)
	@mkdir -p api_docs/dist
	@docker run --rm -v $(PWD)/api_docs:/spec redocly/cli bundle /spec/openapi.yaml --output /spec/dist/openapi.yaml

validate-docs: ## Validate OpenAPI spec (requires Docker)
	@docker run --rm -v $(PWD)/api_docs:/spec redocly/cli lint /spec/openapi.yaml

docs-up: compile-docs ## Compile OpenAPI docs and start docs server at http://localhost:8081
	docker compose --profile docs up docs -d

docs-down: ## Stop API docs server
	docker compose --profile docs down docs

seed: ## Seed URLs into the API (flags: n=100 workers=10 target=http://localhost:8000)
	go run ./cmd/seed -n $(or $(n),100) -workers $(or $(workers),10) -target $(or $(target),http://localhost:8000)

validate-krakend: ## Validate KrakenD configuration (requires Docker)
	@mkdir -p krakend/dist
	@docker run --rm \
		-v $(PWD)/krakend:/etc/krakend \
		-e FC_ENABLE=1 \
		-e FC_SETTINGS=/etc/krakend/settings/dev \
		-e FC_PARTIALS=/etc/krakend/partials \
		-e FC_TEMPLATES=/etc/krakend/templates \
		-e FC_OUT=/etc/krakend/dist/compiled.json \
		krakend:$(KRAKEND_VERSION) check -c /etc/krakend/krakend.tmpl.json; \
	grep -n '<' krakend/dist/compiled.json 2>/dev/null || true
