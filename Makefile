PROJECT_NAME := "gbc"
PKG := "gitlab.com/MoBlaa/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ./... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)

.PHONY: all dep build clean test coverage coverhtml lint bench

all: test

lint: dep
	@golint -set_exit_status ${PKG_LIST}

test:
	@go test --short ${PKG_LIST}

race: dep ## Run data race detector
	@go test -race --short ${PKG_LIST}

coverage: ## Generate global code coverage report
	@mkdir -p cover
	@go test -coverprofile=cover/coverage.out ${PKG_LIST}
	@go tool cover -html=cover/coverage.out -o cover/coverage.html
	@go tool cover -func cover/coverage.out

dep: ## Get the dependencies
	@go get -v -d ./...
	@go get golang.org/x/lint/golint

clean: ## Remove previous build
	@rm -fr bin
	@find . -type d -name logs -prune -exec rm -rf {} \;
