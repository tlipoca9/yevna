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
	command -v ginkgo && ginkgo run --label-filter=!benchmark -cover -coverprofile=cover.out ./... || go test -cover ./...

.PHONY: bench
bench:
	command -v ginkgo && ginkgo run --label-filter=benchmark ./... || go test -bench=. ./...
