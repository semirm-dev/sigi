# A thin wrapper: the build lives in build.zig, and these only save typing.
ZIG     ?= zig
BIN     := bin
TARGETS := x86_64-linux aarch64-linux x86_64-windows aarch64-windows
# Zig only finds the macOS SDK for native builds, so a Mac adds just itself.
ifeq ($(shell uname -s),Darwin)
TARGETS += native
endif

.PHONY: help build build-linux build-win build-osx run test lint release tag clean

help: ## Show available targets
	@grep -E '^[a-z][a-z-]*:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-12s %s\n", $$1, $$2}'

build: ## Build zig-out/bin/sigi for this machine
	$(ZIG) build

build-linux: ## ReleaseSmall bin/sigi-linux-amd64
	$(ZIG) build -Doptimize=ReleaseSmall -Dtarget=x86_64-linux --prefix zig-out/build-linux
	@mkdir -p $(BIN) && cp zig-out/build-linux/bin/sigi $(BIN)/sigi-linux-amd64
	@echo "  $(BIN)/sigi-linux-amd64"

build-win: ## ReleaseSmall bin/sigi-windows-amd64.exe
	$(ZIG) build -Doptimize=ReleaseSmall -Dtarget=x86_64-windows --prefix zig-out/build-win
	@mkdir -p $(BIN) && cp zig-out/build-win/bin/sigi.exe $(BIN)/sigi-windows-amd64.exe
	@echo "  $(BIN)/sigi-windows-amd64.exe"

# Native only: Zig finds Apple's frameworks for a native build and nowhere
# else, so cross-compiling this one dies on -framework CoreGraphics.
build-osx: ## ReleaseSmall bin/sigi-macos-<arch> (on a Mac; cannot cross-compile)
	@[ "$$(uname -s)" = "Darwin" ] || { echo "make build-osx needs a Mac: Zig only finds the frameworks for a native build"; exit 1; }
	$(ZIG) build -Doptimize=ReleaseSmall --prefix zig-out/build-osx
	@mkdir -p $(BIN)
	@arch=$$(uname -m); case $$arch in x86_64) arch=amd64 ;; esac; \
		cp zig-out/build-osx/bin/sigi $(BIN)/sigi-macos-$$arch; \
		echo "  $(BIN)/sigi-macos-$$arch"

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

# Bump, commit and tag in one step. Doing these by hand leaves a gap between
# the bump and the tag where a stale HEAD gets tagged instead, which fails the
# release job -- the version it finds is the old one.
tag: ## Bump .version, commit it and tag (make tag VERSION=2.0.3)
	@[ -n "$(VERSION)" ] || { echo "usage: make tag VERSION=2.0.3"; exit 1; }
	@git diff --quiet && git diff --cached --quiet \
		|| { echo "working tree is dirty; commit or stash first"; exit 1; }
	@sed -i.bak 's/\.version = "[^"]*"/.version = "$(VERSION)"/' build.zig.zon && rm -f build.zig.zon.bak
	@grep -q '\.version = "$(VERSION)"' build.zig.zon || { echo "could not set .version"; exit 1; }
	@$(ZIG) build
	@test "$$(./zig-out/bin/sigi --version)" = "sigi $(VERSION)" \
		|| { echo "built binary does not report $(VERSION)"; exit 1; }
	@git add build.zig.zon && git commit -q -m "Bump to $(VERSION)"
	@git tag -a v$(VERSION) -m "sigi $(VERSION)"
	@echo "  tagged v$(VERSION) on $$(git rev-parse --short HEAD)"
	@echo "  push with: git push && git push origin v$(VERSION)"

clean: ## Remove build output
	rm -rf zig-out .zig-cache $(BIN)
