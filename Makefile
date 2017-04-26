NAME := gocrawsan
COMMAND_NAME := goc
SRCS := $(shell find . -type d -name vendor -prune -o -type f -name "*.go" -print)
VERSION := 0.1.0
REVISION := $(shell git rev-parse --short HEAD)
LDFLAGS := -ldflags="-s -w -X \"main.Version=$(VERSION)\" -X \"main.Revision=$(REVISION)\"" 
DIST_DIRS := find * -type d -exec
SHA256_386 = $(shell cat dist/$(NAME)-$(VERSION)-darwin-386.tar.gz | openssl dgst -sha256 | sed 's/^.* //')
SHA256_AMD64 = $(shell cat dist/$(NAME)-$(VERSION)-darwin-amd64.tar.gz | openssl dgst -sha256 | sed 's/^.* //')

.DEFAULT_GOAL := bin/$(COMMAND_NAME)

.PHONY: test
test: glide
	@go test -cover -v `glide novendor`

.PHONY: install
install: build
	@go install ./cmd

.PHONY: uninstall
uninstall:

.PHONY: clean
clean:
	@rm -rf bin/*
	@rm -rf vendor/*
	@rm -rf dist/*

.PHONY: dist-clean
dist-clean: clean
	@rm -f $(NAME).tar.gz

.PHONY: build
build:
	-@goimports -w .
	@gofmt -w .
	@go build $(LDFLAGS)

.PHONY: cross-build
cross-build: deps
	@for os in darwin linux windows; do \
	    for arch in amd64 386; do \
	        GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -a -tags netgo \
	        -installsuffix netgo $(LDFLAGS) -o dist/$$os-$$arch/$(COMMAND_NAME); \
	    done; \
	done

.PHONY: glide
glide:
ifeq ($(shell command -v glide 2> /dev/null),)
	curl https://glide.sh/get | sh
endif

.PHONY: deps
deps: glide
	glide install

.PHONY: bin/$(COMMAND_NAME) 
bin/$(COMMAND_NAME): $(SRCS)
	go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o bin/$(COMMAND_NAME) 

.PHONY: dist
dist:
	@cd dist && \
	$(DIST_DIRS) cp ../LICENSE {} \; && \
	$(DIST_DIRS) cp ../README.md {} \; && \
	$(DIST_DIRS) cp ../completions/zsh/_$(COMMAND_NAME) {} \; && \
	$(DIST_DIRS) tar zcf $(NAME)-$(VERSION)-{}.tar.gz {} \;

