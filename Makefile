GO ?= go
GOCACHE ?= /tmp/uclaw-gocache

.PHONY: fmt build test verify tidy desktop phase0

fmt:
	gofmt -w $$(find cli core internal -name '*.go' 2>/dev/null)

tidy:
	$(GO) mod tidy

build:
	GOCACHE=$(GOCACHE) $(GO) build ./...

test:
	GOCACHE=$(GOCACHE) $(GO) test ./...

verify: fmt build test

desktop:
	node desktop/build.js

phase0: tidy verify
	@echo "Phase 0 complete."
