# Pollen Makefile
#
# FORK-03 (Reproducible builds): the `verify-release` target performs two
# independent clean builds of the same source and asserts byte-for-byte
# identical sha256 hashes. A safety harness that cannot prove its own build
# integrity is not trustworthy, so this gate ships from v0.1.1-pollen.1.
#
# Reproducibility requirements (RESEARCH Pitfall 1):
#   - The SAME Go toolchain version MUST be used for both builds. The toolchain
#     is pinned via the `toolchain` directive in go.mod (go 1.25). Building with
#     a different toolchain WILL produce a different hash even from identical
#     source — that is expected, not a reproducibility failure.
#   - ldflags inject ONLY the commit DATE (`git show -s --format=%cI HEAD`),
#     never a wall-clock shell timestamp, which would differ on every build and
#     break reproducibility.
#   - `-trimpath` strips local filesystem paths; `-buildvcs=false` suppresses
#     Go 1.18+ embedded VCS stamping; `-mod=readonly` forbids implicit go.mod
#     mutation during the build.
#   - Pollen keeps `Version` in package main (not an internal/version package),
#     so the ldflag path is simply `main.Version` — no version package to inject.

# Module path + version variable location (package main in cmd/pollen/version.go).
MODULE     := github.com/bantuson/pollen
VERSION    ?= dev
COMMIT     := $(shell git rev-parse HEAD)
# Commit DATE (RFC-3339), NOT the wall-clock build time — required for
# reproducibility. Never replace this with a shell wall-clock timestamp.
DATE       := $(shell git show -s --format=%cI HEAD)

# Reproducible build flags (CLAUDE.md Build Constraints / RESEARCH Pattern 4).
GOFLAGS    := -trimpath -buildvcs=false -mod=readonly
# NOTE: Pollen uses main.Version (package main), NOT a version subpackage.
# Do NOT inject Commit or Date here — pollen has no internal/version package.
LDFLAGS    := -s -w -X main.Version=$(VERSION)

.PHONY: build test vet verify-release

build:
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o dist/pollen ./cmd/pollen

test:
	go test -race -count=1 ./...

vet:
	go vet ./...

# verify-release (FORK-03): prove the release build is byte-for-byte
# reproducible. Builds the same source TWICE into separate output paths with
# identical flags, computes sha256 of each artifact, and exits non-zero if the
# hashes differ. Requires VERSION (e.g. `make verify-release VERSION=0.1.1-pollen.1`).
#
# sha256sum is used where available (Linux, macOS via coreutils, Git Bash on
# Windows). Portable fallback: `shasum -a 256` (BSD/macOS default) — uncomment
# the fallback line below if sha256sum is absent on your platform.
verify-release:
	@if [ -z "$(VERSION)" ] || [ "$(VERSION)" = "dev" ]; then \
		echo "ERROR: verify-release requires an explicit VERSION (e.g. make verify-release VERSION=0.1.1-pollen.1)"; \
		exit 1; \
	fi
	@echo ">> Reproducibility check for VERSION=$(VERSION)"
	@echo ">> Go toolchain in use (must be identical for both builds):"
	@go version
	@rm -rf dist/verify-a dist/verify-b
	@mkdir -p dist/verify-a dist/verify-b
	go build $(GOFLAGS) -ldflags "-s -w -X main.Version=$(VERSION)" \
		-o dist/verify-a/pollen ./cmd/pollen
	go build $(GOFLAGS) -ldflags "-s -w -X main.Version=$(VERSION)" \
		-o dist/verify-b/pollen ./cmd/pollen
	@HASH_A=$$(sha256sum dist/verify-a/pollen | cut -d' ' -f1); \
	HASH_B=$$(sha256sum dist/verify-b/pollen | cut -d' ' -f1); \
	echo "build A: $$HASH_A"; \
	echo "build B: $$HASH_B"; \
	if [ "$$HASH_A" != "$$HASH_B" ]; then \
		echo "FAIL: builds are NOT byte-for-byte reproducible (hash mismatch)"; \
		exit 1; \
	fi; \
	echo "OK: builds are byte-for-byte reproducible (FORK-03)"
