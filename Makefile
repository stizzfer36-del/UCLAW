GO ?= go
GOCACHE ?= /tmp/uclaw-gocache

.PHONY: fmt build test verify security desktop phase0

fmt:
	gofmt -w $(shell find cmd internal -name '*.go' -print)

build:
	GOCACHE=$(GOCACHE) $(GO) build ./...

test:
	GOCACHE=$(GOCACHE) $(GO) test ./...

verify: fmt build test

desktop:
	node desktop/build.js

security: verify desktop
	GOCACHE=$(GOCACHE) $(GO) test ./internal/hardening ./internal/voice ./internal/desktop ./internal/app

phase0: verify
