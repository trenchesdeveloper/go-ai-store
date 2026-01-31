.PHONY: help build run dev lint fix migrate-up migrate-down migratecreate docker-up docker-down sqlc docs-generate

help:
	@echo "Available targets:"
	@echo "  build    Build the API binary"
	@echo "  run      Run the API binary"
	@echo "  dev      Run the API binary with race detection and dev tag"
	@echo "  lint     Run linter on the codebase"
	@echo "  migrate-up    Run database migrations"
	@echo "  migrate-down    Rollback database migrations"
	@echo "  docker-up    Start docker containers"
	@echo "  docker-down    Stop docker containers"
	@echo "  sqlc    Generate database queries"
	@echo "  docs-generate    Generate Swagger documentation"

build:
	@echo "Building all binaries..."
	@mkdir -p bin
	@for cmd in ./cmd/*/; do \
		if [ -d "$$cmd" ] && [ -f "$$cmd/main.go" ]; then \
			binary=$$(basename "$$cmd"); \
			echo "Building $$binary..."; \
			go build -o "bin/$$binary" "$$cmd"; \
		fi; \
	done
	@echo "Build complete!"

run:
	go run cmd/api/main.go

dev:
	go run -race -tags dev cmd/api/main.go

lint:
	golangci-lint run ./cmd/... ./internal/... ./db/...

fix:
	golangci-lint run --fix ./cmd/... ./internal/... ./db/...

migrate-up:
	migrate -path db/migrations -database "postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable" up

migrate-down:
	migrate -path db/migrations -database "postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable" down

migratecreate:
	migrate create -ext sql -dir db/migrations -seq $(name)

docker-up:
	docker compose -f docker/docker-compose.yml up -d

docker-down:
	docker compose -f docker/docker-compose.yml down

sqlc:
	sqlc generate

docs-generate:
	swag init -g cmd/api/main.go -o docs
