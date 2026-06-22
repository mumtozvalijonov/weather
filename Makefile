.PHONY: build up down logs test-backend test-web

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

test-web:
	docker run --rm -e CI=true -v "$$(pwd)/web:/app" -w /app node:24-alpine sh -c "corepack enable && corepack prepare pnpm@10.26.0 --activate && pnpm install --frozen-lockfile && pnpm lint"
