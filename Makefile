BINARY := rest-api-fuzzer

.PHONY: test build run-example clean

test:
	go test ./...

build:
	go build -o dist/$(BINARY) ./cmd/rest-api-fuzzer

run-example:
	go run ./cmd/rest-api-fuzzer -spec examples/openapi.yaml -base-url http://127.0.0.1:8080 -cases 5 -seed 2026

clean:
	rm -rf dist coverage.out
