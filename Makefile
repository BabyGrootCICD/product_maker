.PHONY: test build run dry-run bench bench-baseline resolution

test:
	go test ./...

build:
	go build -o bin/discovery ./cmd/discovery

run:
	go run ./cmd/discovery run --config configs/pipeline.yaml

dry-run:
	go run ./cmd/discovery run --config configs/pipeline.yaml --dry-run

bench:
	go test -bench=. -benchmem -run='^$' ./...

bench-baseline:
	go test -bench=. -benchmem -count=5 -run='^$' ./... | tee out/benchmark-baseline.txt

resolution:
	./scripts/bench-resolution.sh
