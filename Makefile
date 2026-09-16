# A thin wrapper: the build lives in build.zig, and these only save typing.
ZIG     ?= zig
BIN     := bin
TARGETS := x86_64-linux aarch64-linux x86_64-windows aarch64-windows
# Zig only finds the macOS SDK for native builds, so a Mac adds just itself.
ifeq ($(shell uname -s),Darwin)
TARGETS += native
endif

.PHONY: help build build-linux build-win build-osx run test lint release clean

help: ## Show available targets
	@grep -E '^[a-z][a-z-]*:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-12s %s\n", $$1, $$2}'

build: ## Build zig-out/bin/sigi for this machine
	$(ZIG) build

build-linux: ## ReleaseSmall bin/sigi-linux (x86_64)
	$(ZIG) build -Doptimize=ReleaseSmall -Dtarget=x86_64-linux --prefix zig-out/build-linux
	@mkdir -p $(BIN) && cp zig-out/build-linux/bin/sigi $(BIN)/sigi-linux
	@echo "  $(BIN)/sigi-linux"

build-win: ## ReleaseSmall bin/sigi-windows.exe (x86_64)
	$(ZIG) build -Doptimize=ReleaseSmall -Dtarget=x86_64-windows --prefix zig-out/build-win
	@mkdir -p $(BIN) && cp zig-out/build-win/bin/sigi.exe $(BIN)/sigi-windows.exe
	@echo "  $(BIN)/sigi-windows.exe"

# Native only: Zig finds Apple's frameworks for a native build and nowhere
# else, so cross-compiling this one dies on -framework CoreGraphics.
build-osx: ## ReleaseSmall bin/sigi-macos (on a Mac; cannot cross-compile)
	@[ "$$(uname -s)" = "Darwin" ] || { echo "make build-osx needs a Mac: Zig only finds the frameworks for a native build"; exit 1; }
	$(ZIG) build -Doptimize=ReleaseSmall --prefix zig-out/build-osx
	@mkdir -p $(BIN) && cp zig-out/build-osx/bin/sigi $(BIN)/sigi-macos
	@echo "  $(BIN)/sigi-macos"

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
	rm -rf zig-out .zig-cache $(BIN)
