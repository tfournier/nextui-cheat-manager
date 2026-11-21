# Cheat Manager - NextUI pak build
#
# The ARM64 build needs the cross toolchain from scripts/setup-toolchain.sh.
# On any machine with Docker, prefix a target with "docker-" to run it inside
# the build image, e.g.:  make docker-release
#
# Main targets:
#   make test            run unit tests (no toolchain needed)
#   make lint            gofmt + go vet
#   make build           cross-compile the binary for tg5040 (ARM64)
#   make pak             assemble the pak folder in build/pak
#   make release         build the release zip in dist/
#   make docker-<target> run <target> inside the build image
#   make vX.Y.Z          release a version: bump pak.json, commit, tag and push

APP          := cheat-manager
PAK_NAME     := Cheat Manager
PLATFORM     := tg5040

PAK_VERSION  := $(shell sed -n 's/.*"version": *"\([^"]*\)".*/\1/p' pak.json | head -n 1)
RELEASE_FILE := $(shell sed -n 's/.*"release_filename": *"\([^"]*\)".*/\1/p' pak.json | head -n 1)
TAG          ?=
VERSION      ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo $(PAK_VERSION))

BUILD_DIR    := build
DIST_DIR     := dist
PAK_DIR      := $(BUILD_DIR)/pak
BIN          := $(BUILD_DIR)/$(APP)
RELEASE_ZIP  := $(DIST_DIR)/$(RELEASE_FILE)

# Cross-compilation settings (Debian 11 multiarch toolchain)
CROSS_ENV    := CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc \
                PKG_CONFIG_PATH=/usr/lib/aarch64-linux-gnu/pkgconfig
SDL_GFX_LIB  := /usr/lib/aarch64-linux-gnu/libSDL2_gfx-1.0.so.0
GO_TAGS      := nodefaultfont
LDFLAGS      := -s -w -X main.version=$(VERSION)

GO_SOURCES   := main.go internal

RELEASE_BRANCH ?= main

DOCKER       ?= docker
DOCKER_IMAGE ?= nextui-cheat-manager-build

.DEFAULT_GOAL := help
.PHONY: help test lint build pak release check-version clean docker-image FORCE

help: ## Show this help
	@grep -E '^[a-zA-Z_%-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "} {printf "  %-18s %s\n", $$1, $$2}'

test: ## Run unit tests (pure Go, no SDL needed)
	CGO_ENABLED=0 go test ./internal/...

lint: ## Check formatting and run go vet
	@test -z "$$(gofmt -l $(GO_SOURCES))" || { echo "Files need gofmt:"; gofmt -l $(GO_SOURCES); exit 1; }
	CGO_ENABLED=0 go vet ./internal/...

build: $(BIN) ## Cross-compile the ARM64 binary

$(BIN): FORCE
	@mkdir -p $(BUILD_DIR)
	$(CROSS_ENV) go vet -tags $(GO_TAGS) .
	$(CROSS_ENV) go build -tags $(GO_TAGS) -trimpath -ldflags "$(LDFLAGS)" -o $@ .

pak: build ## Assemble the pak folder in build/pak
	rm -rf $(PAK_DIR)
	mkdir -p $(PAK_DIR)/resources/lib
	cp pak/launch.sh $(BIN) pak.json README.md LICENSE $(PAK_DIR)/
	cp pak/THIRD_PARTY.md $(PAK_DIR)/resources/
	cp -L $(SDL_GFX_LIB) $(PAK_DIR)/resources/lib/
	chmod +x $(PAK_DIR)/launch.sh $(PAK_DIR)/$(APP)

release: check-version pak ## Build the release zip (Pak Store format) in dist/
	rm -rf $(DIST_DIR)
	mkdir -p $(DIST_DIR)
	# The pak's contents sit at the root of the zip: the Pak Store creates
	# Tools/<PLATFORM>/Cheat Manager.pak/ itself and extracts the zip into it
	# (a top "Cheat Manager.pak/" folder would nest it and the pak would not start).
	cd $(PAK_DIR) && zip -qr "$(CURDIR)/$(RELEASE_ZIP)" .
	@zip -sf "$(RELEASE_ZIP)" | grep -qx '  launch.sh' || { echo "launch.sh is not at the root of $(RELEASE_ZIP)"; exit 1; }
	@ls -l $(DIST_DIR)

check-version: ## Check that TAG (when set, e.g. by the CI) matches pak.json and has a changelog entry
	@if [ -n "$(TAG)" ]; then \
		[ "$(TAG)" = "$(PAK_VERSION)" ] || { echo "Tag $(TAG) does not match pak.json version $(PAK_VERSION)"; exit 1; }; \
		grep -q '^    "$(TAG)":' pak.json || { echo "pak.json has no changelog entry for $(TAG)"; exit 1; }; \
	fi

v%: FORCE ## Release a version, e.g. make v1.2.3 (pak.json must have its changelog entry)
	@set -e; V="$@"; \
	echo "$$V" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+(\.[0-9]+)?$$' || { echo "Invalid version $$V (expected vX.Y.Z)"; exit 1; }; \
	[ "$$(git rev-parse --abbrev-ref HEAD)" = "$(RELEASE_BRANCH)" ] || { echo "Releases are made from $(RELEASE_BRANCH)"; exit 1; }; \
	[ -z "$$(git status --porcelain)" ] || { echo "The working tree has uncommitted changes"; exit 1; }; \
	if git rev-parse -q --verify "refs/tags/$$V" >/dev/null; then echo "Tag $$V already exists"; exit 1; fi; \
	grep -q "\"Unreleased\":" pak.json || grep -q "^    \"$$V\":" pak.json || { echo "Add a changelog entry (Unreleased or $$V) in pak.json first"; exit 1; }
	$(MAKE) lint test
	@set -e; V="$@"; \
	sed "s/\"Unreleased\":/\"$$V\":/" pak.json > pak.json.tmp && mv pak.json.tmp pak.json; \
	sed "s/^\(  \"version\": *\"\)[^\"]*\"/\1$$V\"/" pak.json > pak.json.tmp && mv pak.json.tmp pak.json; \
	if ! git diff --quiet -- pak.json; then git commit -q -m "Release $$V" -- pak.json; fi; \
	git tag -a "$$V" -m "Release $$V"; \
	git push --atomic origin HEAD "refs/tags/$$V"; \
	echo "Released $$V: the CI will publish the GitHub release."

clean: ## Remove build outputs
	rm -rf $(BUILD_DIR) $(DIST_DIR)

docker-image: ## Build the Docker build image
	$(DOCKER) build --platform linux/amd64 -t $(DOCKER_IMAGE) .

docker-%: docker-image ## Run a target inside the build image (e.g. make docker-release)
	$(DOCKER) run --rm --platform linux/amd64 \
		-v "$(CURDIR)":/src -w /src \
		-v nextui-cheat-manager-gomod:/root/go/pkg/mod \
		-v nextui-cheat-manager-gocache:/root/.cache/go-build \
		-e VERSION="$(VERSION)" \
		$(DOCKER_IMAGE) make $*

FORCE:
