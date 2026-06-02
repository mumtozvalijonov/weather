.PHONY: dev-backend

dev-backend:
	cd backend && go run cmd/api/main.go
