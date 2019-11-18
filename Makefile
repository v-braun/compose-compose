PROJECTNAME := $(shell basename "$(PWD)")
GOBASE := $(shell pwd)
GOPATH := $(GOBASE)/vendor:$(GOBASE)
GOBIN := $(GOBASE)/bin
GOFILES := $(wildcard *.go)

CLI_ENTRY := $(GOBASE)
CLI_BIN_NAME := compose-compose



ci: go-compile go-test-cover

run: go-build exec-bin

exec-bin:
	$(GOBIN)/$(CLI_BIN_NAME)

go-build:
	@echo "  ⚙️  Building binary..."
	@go build -o $(GOBIN)/$(CLI_BIN_NAME) $(CLI_ENTRY)


go-get:
	@echo "  🔎  Checking if there is any missing dependencies..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go get $(get)

go-clean:
	@echo "  🗑  Cleaning build cache"
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go clean

go-test-cover:
	@go test ./... -coverpkg=./... -coverprofile=coverage.txt  -timeout 30s
	@go tool cover -func=coverage.txt

go-compile: go-get go-build

