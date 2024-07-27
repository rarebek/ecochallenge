DB_URL := "postgres://postgres:mubina2007@localhost:5432/ecochallengedb?sslmode=disable"

run:
	go run cmd/main.go

swag-gen:
	swag init -g api/router.go -o api/docs

create-migration:
	migrate create -ext sql -dir migrations -seq tables_up

migrate-up:
	migrate -path migrations -database "${DB_URL}" -verbose up

migrate-down:
	migrate -path migrations -database "${DB_URL}" -verbose down

migration-version:
	migrate -database ${DB_URL} -path migrations version 

migrate-dirty:
	migrate -path ./migrations/ -database ${DB_URL} force "$(number)"