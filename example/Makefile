#!/usr/bin/env make -f
TOPDIR := $(realpath $(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
SELF := $(abspath $(lastword $(MAKEFILE_LIST)))

GIT_URL := $(shell git remote -v|grep push|awk '{print $$2}')
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
GIT_COMMIT := $(shell git rev-parse HEAD)
BUILD_TIME := $(shell date +"%Y-%m-%d %H:%M:%S %Z")
LDFLAGS := "-w -s"
ARCH := $(shell uname -m)
BINARY_NAME := demo

COMPRESS_GO_BINARIES ?= 1

ifeq ($(COMPRESS_GO_BINARIES),1)
ifndef UPX
ifeq ($(shell which upx >/dev/null 2>&1 && echo OK),OK)
UPX = upx
endif
endif
UPX ?= $(error upx not found in PATH. Please install upx or set COMPRESS_GO_BINARIES=0)
endif

.PHONY: help
help: ## Show help message (list targets)
	@awk 'BEGIN {FS = ":.*##"; printf "\nTargets:\n"} /^[$$()% 0-9a-zA-Z_-]+:.*?##/ {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(SELF)

show-var-%:
	@{ \
	escaped_v="$(subst ",\",$($*))" ; \
	if [ -n "$$escaped_v" ]; then v="$$escaped_v"; else v="(undefined)"; fi; \
	printf "%-19s %s\n" "$*" "$$v"; \
	}

SHOW_ENV_VARS = \
	TOPDIR \
	SELF \
	GIT_URL \
	GIT_BRANCH \
	GIT_COMMIT \
	BUILD_TIME \
	LDFLAGS \
	ARCH \
	BINARY_NAME \
	COMPRESS_GO_BINARIES

show-env: $(addprefix show-var-, $(SHOW_ENV_VARS)) ## Show environment details

.PHONY: build
build: clean ## Build binaries on Linux host
	@{ \
	echo "Building binaries..."; \
	set -e ; \
	cd $(TOPDIR); \
	mkdir -p $(TOPDIR)/bin; \
	go build -trimpath -gcflags=all='-l' -ldflags $(LDFLAGS) -o $(TOPDIR)/bin/$(BINARY_NAME)_$(ARCH) $(TOPDIR)/cmd/... ; \
	strip --strip-unneeded $(TOPDIR)/bin/$(BINARY_NAME)_* ; \
	}
ifeq ($(COMPRESS_GO_BINARIES),1)
	@echo "Compressing binary with UPX..."
	$(UPX) --force-overwrite --lzma --quiet --best $(TOPDIR)/bin/$(BINARY_NAME)_*
	@echo "Compression completed!"
endif

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf $(TOPDIR)/bin/
	@echo "Clean completed!"

