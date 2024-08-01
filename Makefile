.PHONY: run
run:
	go run ./cmd/web

# Quality Control

.PHONY: audit
audit: vendor
	@echo "formatting code"
	go fmt ./...
	@echo "vetting code..."
	go vet ./...
	staticcheck ./...
	@echo "running tests..."
	go test -race -vet=off ./...

.PHONY: vendor
vendor:
	@echo "tidying..."
	go mod tidy
	@echo "verifying..."
	go mod verify
	@echo "vedoring..."
	go mod vendor

.PHONY: build
build:
	echo "building ..."
	go build -ldflags='-s' -o=./bin/app ./cmd/web
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/app ./cmd/web
