.PHONY: dev build clean migrate-up migrate-down test

dev:
	docker-compose up --build

build:
	docker-compose build

clean:
	docker-compose down -v
	docker system prune -f

migrate-up:
	cd apps/api && go run cmd/migrate/main.go up

migrate-down:
	cd apps/api && go run cmd/migrate/main.go down

test:
	cd apps/api && go test ./...

stop:
	docker-compose down



