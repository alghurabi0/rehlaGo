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

#####
# BUILD
####
git_description = $(shell git describe --always --dirty --tags --long)
linker_flags = '-s -X main.version=${git_description}'

.PHONY: build
build:
	echo "building ..."
	go build -ldflags=${linker_flags} -o=./bin/app ./cmd/web
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o=./bin/linux_amd64/app ./cmd/web
