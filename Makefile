.PHONY: help test test-v test-cover cover-html lint vet fmt tidy doc clean

## ─── Help ────────────────────────────────────────────────────────────────────

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

## ─── Testing ─────────────────────────────────────────────────────────────────

test: ## Run all tests
	go test ./... -count=1

test-v: ## Run all tests (verbose)
	go test ./... -v -count=1

test-cover: ## Run tests with coverage summary
	go test ./... -cover -count=1

cover-html: ## Generate HTML coverage report and open it
	go test ./... -coverprofile=coverage.out -count=1
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html

## ─── Code Quality ────────────────────────────────────────────────────────────

lint: vet ## Run linters (vet + staticcheck if installed)
	@which staticcheck > /dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed, skipping (go install honnef.co/go/tools/cmd/staticcheck@latest)"

vet: ## Run go vet
	go vet ./...

fmt: ## Format all Go files
	gofmt -s -w .

## ─── Dependencies ────────────────────────────────────────────────────────────

tidy: ## Tidy go.mod and go.sum
	go mod tidy

## ─── Documentation ───────────────────────────────────────────────────────────

doc: ## Start local pkgsite documentation server
	@which pkgsite > /dev/null 2>&1 || go install golang.org/x/pkgsite/cmd/pkgsite@latest
	pkgsite -open .

## ─── Cleanup ─────────────────────────────────────────────────────────────────

clean: ## Remove generated files
	rm -f coverage.out coverage.html
