include .envrc

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	go run ./cmd/api -db-dsn=${GO_TEMPLATE_API_DB_DSN}



## run/api/smtp: run the cmd/api application
.PHONY: run/api/mail
run/api/mail:
	go run ./cmd/api -db-dsn=${GO_TEMPLATE_API_DB_DSN} -smtp-username=${MAIL_TEST_USERNAME} -smtp-password=${MAIL_TEST_PASSWORD}



## run/api: run the cmd/api application
.PHONY: run/api/cors
run/api/cors:
	go run ./cmd/api -cors-trusted-origins="http://localhost:3000" -db-dsn=${GO_TEMPLATE_API_DB_DSN}


## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${GO_TEMPLATE_API_DB_DSN}

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database=${GO_TEMPLATE_API_DB_DSN} up

## db/migrations/up: apply all down database migrations
.PHONY: db/migrations/down
db/migrations/down: confirm
	@echo 'Running down migrations...'
	migrate -path ./migrations -database=${GO_TEMPLATE_API_DB_DSN} down
# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: format all .go files and tidy module dependencies
.PHONY: tidy
tidy:
	@echo 'Formatting .go files...'
	go fmt ./...
	@echo 'Tidying module dependencies...'
	go mod tidy
	go mod verify
	go mod vendor

## audit: run quality control checks
.PHONY: audit
audit:
	@echo 'Checking module dependencies'
	go mod tidy -diff
	go mod verify
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...