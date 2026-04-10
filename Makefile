GO ?= go
GOCACHE ?= /tmp/uclaw-gocache

.PHONY: fmt build test verify tidy desktop phase0 bootstrap dev claude

# One-command setup (Claude Code style)
bootstrap:
	@bash scripts/bootstrap.sh

# Quick dev loop: tidy + build + run status
dev:
	@export GOPROXY=$${GOPROXY:-direct}; \
	$(GO) mod tidy && \
	$(GO) build -o ./uclaw ./cli && \
	./uclaw status

# OpenClaw-style: launch with Claude as the default provider
claude:
	@if [ ! -x ./uclaw ]; then $(MAKE) dev; fi
	UCLAW_LLM_PROVIDER=anthropic UCLAW_PROFILE=claude-operator ./uclaw desktop --html

fmt:
	gofmt -w $$(find cli core internal -name '*.go' 2>/dev/null)

tidy:
	$(GO) mod tidy

build:
	GOCACHE=$(GOCACHE) $(GO) build -o ./uclaw ./cli

test:
	GOCACHE=$(GOCACHE) $(GO) test ./...

verify: fmt build test

desktop:
	node desktop/build.js

phase0: tidy verify
	@echo "Phase 0 complete."
