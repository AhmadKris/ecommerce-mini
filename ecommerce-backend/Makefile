.PHONY: run test test-integration test-coverage lint govulncheck migrate-up migrate-down migrate-create docker-build docker-up docker-down

run:
	go run ./cmd/api

test:
	go test ./...

# Spins up real Postgres containers via testcontainers-go — needs a working
# Docker daemon, takes longer than `test`, and is not part of the default
# suite. Covers what curl-against-Docker used to verify manually: the
# checkout race condition and cross-user cart ownership.
test-integration:
	go test -tags=integration ./...

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

lint:
	golangci-lint run

# Same check backend-ci.yml's govulncheck job runs — run this before pushing
# so a reachable CVE surfaces locally, not first as a red CI check.
govulncheck:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

migrate-up:
	migrate -path migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path migrations -database "$$DATABASE_URL" down 1

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down
