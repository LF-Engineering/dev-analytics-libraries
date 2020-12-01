PKG_LIST := $(shell go list ./... | grep -v /vendor/)

lint: ## Lint the files
	golint -set_exit_status $(shell go list ./... | grep -v /vendor/)

test:
	go test ./... -v | grep -v /vendor/

build:
	go build ./...
