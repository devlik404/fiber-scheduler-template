.PHONY: run test build tidy generate-job

run:
	go run ./cmd/api

test:
	go test ./...

build:
	go build -o bin/app ./cmd/api

tidy:
	go mod tidy

generate-job:
	go run ./cmd/generate-job --name "$(NAME)" --schedule "$(SCHEDULE)"
