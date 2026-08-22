.PHONY: test build run dry-run

test:
	go test ./...

build:
	go build -o bin/discovery ./cmd/discovery

run:
	go run ./cmd/discovery run --config configs/pipeline.yaml

dry-run:
	go run ./cmd/discovery run --config configs/pipeline.yaml --dry-run
