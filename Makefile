.PHONY: run test cover build

run:
	go run ./cmd/api

test:
	go test ./...

cover:
	mkdir -p coverage
	go test ./... -coverprofile=coverage/coverage.out
	go tool cover -func=coverage/coverage.out | tee coverage/summary.txt

build:
	CGO_ENABLED=0 go build -o bin/api ./cmd/api
