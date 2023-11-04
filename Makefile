OS=linux
ARCH=amd64

test:
	go test -cover -race -timeout=120s -count 1 ./...

dev-run: 
	go run ./cmd/main.go

db-run:
	docker compose --profile db up -d

app-run:
	docker compose --profile app up

app-run-build:
	docker compose --profile app up --build

down:
	docker compose down

down-and-clear-all:
	docker compose down --remove-orphans
