# A thin wrapper: the build lives in build.zig, and these only save typing.
ZIG     ?= zig
BIN     := bin
# <published name>:<zig target>. Published names read <os>-<arch> with the
# words people recognise; Zig's own triples read the other way round. macOS is
# not in the list because Zig finds Apple's frameworks only for a native build,
# so a Mac adds itself and nothing else can.
TARGETS := linux-amd64:x86_64-linux linux-arm64:aarch64-linux \
           windows-amd64:x86_64-windows windows-arm64:aarch64-windows
MACARCH := $(shell uname -m | sed -e s/x86_64/amd64/)
ifeq ($(shell uname -s),Darwin)
TARGETS += macos-$(MACARCH):native
endif
# macOS has shasum instead.
SHASUM := $(shell command -v sha256sum >/dev/null 2>&1 && echo sha256sum || echo shasum -a 256)

.PHONY: help build build-linux build-win build-osx run test lint release tag clean

help: ## Show available targets
	@grep -E '^[a-z][a-z-]*:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-12s %s\n", $$1, $$2}'

build: ## Debug build of zig-out/bin/sigi, for working on sigi
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
	@cp zig-out/build-osx/bin/sigi $(BIN)/sigi-macos-$(MACARCH)
	@echo "  $(BIN)/sigi-macos-$(MACARCH)"

run: ## Run sigi; pass flags with ARGS="-i 30s -v"
	$(ZIG) build run -- $(ARGS)

test: ## Run the unit tests
	$(ZIG) build test --summary all

lint: ## Check formatting
	$(ZIG) fmt --check build.zig build.zig.zon src

# What a tag publishes, built here: same names, same flags, plus checksums.
# On anything but a Mac that is every platform except macOS.
release: ## ReleaseSmall bin/sigi-<os>-<arch> for every target, plus SHA256SUMS
	@rm -rf $(BIN) && mkdir -p $(BIN)
	@set -e; for t in $(TARGETS); do \
		name=$${t%%:*}; target=$${t##*:}; \
		ext=""; case $$name in windows-*) ext=.exe ;; esac; \
		$(ZIG) build -Doptimize=ReleaseSmall -Dtarget=$$target --prefix zig-out/release/$$name; \
		cp zig-out/release/$$name/bin/sigi$$ext $(BIN)/sigi-$$name$$ext; \
		echo "  $(BIN)/sigi-$$name$$ext"; \
	done
	@cd $(BIN) && $(SHASUM) sigi-* > SHA256SUMS && echo "  $(BIN)/SHA256SUMS"

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
