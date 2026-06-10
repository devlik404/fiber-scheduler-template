.PHONY: run test build tidy generate-job generate-api

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

generate-api:
	go run ./cmd/generate-api --name "$(NAME)" --method "$(METHOD)" --path "$(PATH)"
