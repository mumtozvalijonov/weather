.PHONY: build up down logs test-backend

build:
	docker compose build

up:
	docker compose up --build

down:
	docker compose down

logs:
	docker compose logs -f

test-backend:
	docker run --rm -v "$$(pwd)/backend:/app" -w /app golang:1.26.3-alpine go test ./...
