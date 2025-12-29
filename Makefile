VERSION := $(shell \
  tag=$$(git describe --tags --exact-match 2>/dev/null || true); \
  if [ -n "$$tag" ] && echo "$$tag" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$$'; then \
    echo "$$tag" | sed 's/^v//'; \
  else \
    git rev-parse --short HEAD; \
  fi)


.PHONY: lint
## lint: Lint source code
lint:
	@golangci-lint run

.PHONY: fmt
## fmt: Format source code
fmt:
	@go tool gofumpt -w .

.PHONY: test
## test: Run tests
test:
	@go tool gotestsum

.PHONY: coverage
## coverage: Generate test coverage report
coverage:
	@go tool gotestsum -- -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

.PHONY: swagger
## swagger: Gerenete swagger docs
swagger:
	@go tool swag init -g cmd/server/main.go --output docs --parseDependency --parseInternal

.PHONY: help
all: help
# help: show help message
help: Makefile
	@echo
	@echo " Choose a command to run in "$(NAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
