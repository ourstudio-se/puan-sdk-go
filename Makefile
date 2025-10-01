.PHONY: test
test: glpk
	@go test -count=5 -race -cover ./...

.PHONY: lint
lint:
	@golangci-lint run ./...

.PHONY: glpk
glpk:
	@docker compose up -d glpk-api

.PHONY: down
down:
	@docker compose down

.PHONY: gitleaks
gitleaks:
	@gitleaks detect --source . --verbose

.PHONY: vulncheck
vulncheck:
	@govulncheck ./...

security-scan: gitleaks vulncheck
