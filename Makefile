.PHONY: serve
build: lint
	go build -o bin/yevna

.PHONY: lint
lint:
	go generate ./...
	go mod tidy
	golangci-lint run --fix ./...

.PHONY: test
test:
	go test -cover ./...

.PHONY: bench
bench:
	go test -bench=. ./...
