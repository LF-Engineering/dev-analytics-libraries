PKG_LIST = $(shell go list ./... | grep -v /vendor/)

lint: fmt
	golint -set_exit_status $(PKG_LIST)

fmt:
	./scripts/for_each_go_file.sh gofmt -s -w

test:
	go test -v $(PKG_LIST)

build:
	go build ./...
