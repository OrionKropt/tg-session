APP_NAME := tg-session
CMD_DIR  := ./cmd/app
BIN_DIR  := ./bin
BIN_FILE := $(BIN_DIR)/$(APP_NAME)

.PHONY: fmt vet build clean run test lint dev tidy

all: build

docker-compose-up:
	sudo docker compose -f docker-compose.yml --env-file .env up --build

docker-compose-down:
	sudo docker compose -f docker-compose.yml --env-file .env down

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BIN_FILE) $(CMD_DIR)

run: build
	go run $(CMD_DIR)

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	@echo "→ Cleaning binary"
	@rm -rf $(BIN_DIR)
