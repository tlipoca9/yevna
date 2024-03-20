.PHONY: serve
build: lint
	go build -o bin/yevna

.PHONY: lint
lint:
	go generate ./...
	go mod tidy
	golangci-lint run --fix ./...
