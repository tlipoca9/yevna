.PHONY: lint
lint:
	go generate ./...
	go mod tidy
	golines -m 100 -w .
	golangci-lint run --fix ./...

.PHONY: test
test:
	command -v ginkgo && ginkgo run --label-filter=!benchmark -cover -coverprofile=cover.out ./... || go test -cover ./...

.PHONY: bench
bench:
	command -v ginkgo && ginkgo run --label-filter=benchmark ./... || go test -bench=. ./...

.PHONY: install-tools
install-tools:
	go install github.com/segmentio/golines@latest
	go install github.com/onsi/ginkgo/v2/ginkgo
