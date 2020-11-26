PKG_LIST := $(shell go list ./... | grep -v /vendor/)

lint: ## Lint the files
	@golint -set_exit_status $(PKG_LIST)

test:
	go test ./... | grep -v /vendor/ -v