PROJECT_NAME := "GO_ECHO"
VERSION := "0.1.0"
PKG := "."
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)
 
.PHONY: all dep lint vet test test-coverage build clean
 
all: build

dep: ## Get the dependencies
	@go mod download

lint: ## Lint Golang files
	@golangci-lint -c .golangci.yml run

vet: ## Run go vet
	@go vet ${PKG_LIST}

test: ## Run unittests
	@go test -short ${PKG_LIST}

docker: ## Run unittests
	@docker buildx build \
	--push -t dseifert/go-echo:${VERSION} \
	--platform=linux/amd64,linux/arm64,linux/ppc64le,linux/s390x,linux/arm/v7 \
	 .

test-coverage: ## Run tests with coverage
	@go test -short -coverprofile cover.out -covermode=atomic ${PKG_LIST} 
	# @cat cover.out >> coverage.txt

build: dep ## Build the binary file
	@go build -i -o out/go-echo $(PKG)/cmd/main
 
clean: ## Remove previous build
	@rm -f $(PROJECT_NAME)/build
 
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'