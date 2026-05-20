.PHONY: migrate-up migrate-down migrate-force migrate-version migrate-create
include .env

DSN=postgres://user_service_admin:$(PG_PASSWORD)5@localhost:5432/user_service?sslmode=disable
MIGRATION_PATH=./migrations

migrate-up:
	migrate -path $(MIGRATION_PATH) -database "$(DSN)" up

migrate-down:
	migrate -path $(MIGRATION_PATH) -database "$(DSN)" down

migrate-force:
	migrate -path $(MIGRATION_PATH) -database "$(DSN)" force $(force)

migrate-version:
	migrate -path $(MIGRATION_PATH) -database "$(DSN)" version

migrate-create:
	migrate create -ext sql -dir $(MIGRATION_PATH) -seq=false $(name)
