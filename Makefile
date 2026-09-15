# A thin wrapper: the build lives in build.zig, and these only save typing.
ZIG     ?= zig
TARGETS := x86_64-linux aarch64-linux x86_64-windows aarch64-windows
# Zig only finds the macOS SDK for native builds, so a Mac adds just itself.
ifeq ($(shell uname -s),Darwin)
TARGETS += native
endif

.PHONY: help build run test lint release clean

help: ## Show available targets
	@grep -E '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-8s %s\n", $$1, $$2}'

build: ## Build zig-out/bin/sigi for this machine
	$(ZIG) build

run: ## Run sigi; pass flags with ARGS="-i 30s -v"
	$(ZIG) build run -- $(ARGS)

test: ## Run the unit tests
	$(ZIG) build test --summary all

lint: ## Check formatting
	$(ZIG) fmt --check build.zig build.zig.zon src

release: ## ReleaseSmall binaries under zig-out/release/<target>
	@for t in $(TARGETS); do \
		$(ZIG) build -Doptimize=ReleaseSmall -Dtarget=$$t --prefix zig-out/release/$$t || exit 1; \
		echo "  zig-out/release/$$t"; \
	done

clean: ## Remove build output
	rm -rf zig-out .zig-cache
