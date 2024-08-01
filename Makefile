.PHONY: run
run:
	go run ./cmd/web

# Quality Control

.PHONY: audit
audit:
	@echo "tidying and verifying module dependencies..."
	go mod tidy
	go mod verify
	@echo "formatting code"
	go fmt ./...
	@echo "vetting code..."
	go vet ./...
	staticcheck ./...
	@echo "running tests..."
	go test -race -vet=off ./...
