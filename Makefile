PROJECT=middleware-server

BUILD_PATH := $(shell pwd)/.gobuild

PROJECT_PATH := "$(BUILD_PATH)/src/github.com/giantswarm"

BIN := $(PROJECT)

.PHONY=clean run-test get-deps update-deps

GOPATH := $(BUILD_PATH)

SOURCE=$(shell find . -name '*.go')

all: get-deps $(BIN)

clean:
	rm -rf $(BUILD_PATH) $(BIN)

get-deps: .gobuild

.gobuild:
	mkdir -p $(PROJECT_PATH)
	cd "$(PROJECT_PATH)" && ln -s ../../../.. $(PROJECT)

	#
	# Fetch private packages first (so `go get` skips them later)

	#
	# Fetch public dependencies via `go get`
	GOPATH=$(GOPATH) builder go get github.com/dchest/uniuri
	GOPATH=$(GOPATH) builder go get github.com/giantswarm/request-context
	GOPATH=$(GOPATH) builder go get github.com/juju/errgo
	GOPATH=$(GOPATH) builder go get github.com/gorilla/mux
	GOPATH=$(GOPATH) builder go get github.com/gorilla/context
	GOPATH=$(GOPATH) go get -d -v github.com/giantswarm/$(PROJECT)

	#
	# Build test packages (we only want those two, so we use `-d` in go get)
	GOPATH=$(GOPATH) go get -v github.com/onsi/gomega
	GOPATH=$(GOPATH) go get -v github.com/onsi/ginkgo

$(BIN): $(SOURCE)
	GOPATH=$(GOPATH) go build -o $(BIN)

run-tests:
	GOPATH=$(GOPATH) go test ./...

build-examples:
	GOPATH=$(GOPATH) go build -o not-found.example ./example/not-found/
	GOPATH=$(GOPATH) go build -o middleware.example ./example/middleware/
	GOPATH=$(GOPATH) go build -o error.example ./example/error/
	GOPATH=$(GOPATH) go build -o fileserver.example ./example/fileserver/
	GOPATH=$(GOPATH) go build -o welcome.example ./example/welcome/
	GOPATH=$(GOPATH) go build -o mux-cooperation.example ./example/mux-cooperation/
	GOPATH=$(GOPATH) go build -o request-callback.example ./example/request-callback/
	GOPATH=$(GOPATH) go build -o close.example ./example/close/
	GOPATH=$(GOPATH) go build -o healthcheck.example ./example/healthcheck/

fmt:
	gofmt -l -w .
