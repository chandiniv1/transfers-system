DB_DRIVER=postgres
DB_USER=root
DB_PASSWORD=secret
DB_NAME=transfers_db
DB_PORT=5433
DB_HOST=localhost
DB_SOURCE="postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"

POSTGRES_CONTAINER_NAME=postgres12
NETWORK_NAME=transfers-network

## Create the database
createdb:
	docker exec -it $(POSTGRES_CONTAINER_NAME) createdb --username=$(DB_USER) --owner=$(DB_USER) $(DB_NAME)

## Drop the database
dropdb:
	docker exec -it $(POSTGRES_CONTAINER_NAME) dropdb $(DB_NAME)

## Start PostgreSQL in Docker
postgres:
	docker rm -f $(POSTGRES_CONTAINER_NAME) || true
	docker run --name $(POSTGRES_CONTAINER_NAME) \
		--network $(NETWORK_NAME) \
		-p $(DB_PORT):5432 \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-d postgres:12-alpine

## Create network if not exists
network:
	docker network create $(NETWORK_NAME)

## Run migration up (all)
migrateup:
	migrate -path db/migration -database $(DB_SOURCE) -verbose up

## Run migration down (all)
migratedown:
	migrate -path db/migration -database $(DB_SOURCE) -verbose down

## Generate SQLC code
sqlc:
	sqlc generate

mockgen:
	mockgen -package mockdb -destination db/mock/store.go github.com/chandiniv1/transfers-system/db/sqlc Store

## Run tests
test:
	go test -v -cover -short ./...

## Run the server
server:
	go run main.go

## Create a new migration
new_migration:
ifndef name
	$(error name is required. Use `make new_migration name=create_accounts_table`)
endif
	migrate create -ext sql -dir db/migration -seq $(name)

.PHONY: createdb dropdb postgres migrateup migratedown sqlc test server new_migration network mockgen
