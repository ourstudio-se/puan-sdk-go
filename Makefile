.PHONY: test
test:
	@go test -count=5 -race -cover ./...

.PHONY: lint
lint:
	@golangci-lint run ./...
