.PHONY: run docker-build docker-up docker-down migrate test generate

run:
	go run ./cmd/server

docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

migrate:
	go run -mod=mod github.com/golang-migrate/migrate/v4/cmd/migrate \
		-path internal/db/migrations \
		-database "postgres://greenpos:greenpos123@localhost:5432/greenpos?sslmode=disable"

test:
	go test ./...

generate:
	go generate ./...
