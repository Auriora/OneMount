.PHONY: all, build, test, test-init, test-python, unit-test, integration-test, system-test, srpm, rpm, dsc, changes, deb, deb-docker, ubuntu, ubuntu-docker, ubuntu-docker-image, deb-pbuilder, docker-build-image, clean, install, install-system, uninstall, uninstall-system, install-dry-run, install-system-dry-run, uninstall-dry-run, uninstall-system-dry-run, validate-packaging, setup-pbuilder, update-pbuilder, update-imports, docker-test-build, docker-test-unit, docker-test-integration, docker-test-system, docker-test-all, docker-test-coverage, docker-test-shell, docker-test-clean

# auto-calculate software/package versions
VERSION := $(shell grep Version packaging/rpm/onemount.spec | sed 's/Version: *//g')
RELEASE := $(shell grep -oP "Release: *[0-9]+" packaging/rpm/onemount.spec | sed 's/Release: *//g')
DIST := $(shell rpm --eval "%{?dist}" 2> /dev/null || echo 1)
RPM_FULL_VERSION = $(VERSION)-$(RELEASE)$(DIST)

# -Wno-deprecated-declarations is for gotk3, which uses deprecated methods for older
# glib compatibility: https://github.com/gotk3/gotk3/issues/762#issuecomment-919035313
CGO_CFLAGS := CGO_CFLAGS=-Wno-deprecated-declarations

# Build directory structure
BUILD_DIR := build
OUTPUT_DIR := $(BUILD_DIR)/binaries
PACKAGE_DIR := $(BUILD_DIR)/packages
DEB_DIR := $(PACKAGE_DIR)/deb
RPM_DIR := $(PACKAGE_DIR)/rpm
SOURCE_DIR := $(PACKAGE_DIR)/source
DOCKER_DIR := $(BUILD_DIR)/docker
TEMP_DIR := $(BUILD_DIR)/temp

# test-specific variables
GORACE := GORACE="log_path=fusefs_tests.race strip_path_prefix=1"
TEST_TIMEOUT := 10m

all: onemount onemount-launcher

build: all


onemount: $(shell find internal/fs/ -type f) cmd/onemount/main.go
	bash scripts/cgo-helper.sh
	mkdir -p $(OUTPUT_DIR)
	$(CGO_CFLAGS) go build -v \
		-o $(OUTPUT_DIR)/onemount \
		-ldflags="-X github.com/auriora/onemount/cmd/common.commit=$(shell git rev-parse HEAD)" \
		./cmd/onemount


onemount-headless: $(shell find internal/fs/ cmd/common/ -type f) cmd/onemount/main.go
	mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=0 go build -v \
		-o $(OUTPUT_DIR)/onemount-headless \
		-ldflags="-X github.com/auriora/onemount/cmd/common.commit=$(shell git rev-parse HEAD)" \
		./cmd/onemount


onemount-launcher: $(shell find internal/ui/ cmd/common/ -type f) cmd/onemount-launcher/main.go
	mkdir -p $(OUTPUT_DIR)
	$(CGO_CFLAGS) go build -v \
		-o $(OUTPUT_DIR)/onemount-launcher \
		-ldflags="-X github.com/auriora/onemount/cmd/common.commit=$(shell git rev-parse HEAD)" \
		./cmd/onemount-launcher


install: onemount onemount-launcher
	@./scripts/dev build manifest --target makefile --type user --action install | bash


install-system: onemount onemount-launcher
	@./scripts/dev build manifest --target makefile --type system --action install | bash


uninstall:
	@./scripts/dev build manifest --target makefile --type user --action uninstall | bash


uninstall-system:
	@./scripts/dev build manifest --target makefile --type system --action uninstall | bash


# Show what would be installed for user installation (dry run)
install-dry-run: onemount onemount-launcher
	@./scripts/dev build manifest --target makefile --type user --action install --dry-run


# Show what would be installed for system installation (dry run)
install-system-dry-run: onemount onemount-launcher
	@./scripts/dev build manifest --target makefile --type system --action install --dry-run


# Show what would be uninstalled for user installation (dry run)
uninstall-dry-run:
	@./scripts/dev build manifest --target makefile --type user --action uninstall --dry-run


# Show what would be uninstalled for system installation (dry run)
uninstall-system-dry-run:
	@./scripts/dev build manifest --target makefile --type system --action uninstall --dry-run


# Validate packaging requirements
validate-packaging:
	@./scripts/dev build manifest --target makefile --action validate | bash
	@test -f scripts/cgo-helper.sh || (echo "Error: cgo-helper.sh script not found" && exit 1)

# Setup pbuilder environment for building packages (legacy)
setup-pbuilder:
	@echo "Setting up pbuilder environment..."
	@if [ ! -f /var/cache/pbuilder/base.tgz ]; then \
		echo "Creating pbuilder base environment..."; \
		sudo pbuilder create --distribution noble; \
	else \
		echo "Pbuilder base environment already exists."; \
	fi
	@echo "Pbuilder setup complete!"

# Update pbuilder environment (legacy)
update-pbuilder:
	@echo "Updating pbuilder environment..."
	sudo pbuilder update --override-config

# Build Docker image for Debian package building
docker-build-image:
	@echo "Building Docker image for Debian package building..."
	docker build -f packaging/docker/Dockerfile.deb-builder -t onemount-deb-builder .
	@echo "Docker image built successfully!"

# Build Ubuntu packages using Docker
deb-docker: ubuntu-docker
ubuntu-docker: validate-packaging
	@echo "Building Ubuntu packages using Docker..."
	./scripts/dev build deb --docker

# Build Ubuntu Docker image
ubuntu-docker-image:
	@echo "Building Ubuntu Docker image for package building..."
	docker build -t onemount-ubuntu-builder -f packaging/docker/Dockerfile.deb-builder .
	@echo "Ubuntu Docker image built successfully!"

# used to create release tarball for rpmbuild
v$(VERSION).tar.gz: validate-packaging
	mkdir -p $(SOURCE_DIR) $(TEMP_DIR)
	rm -rf $(TEMP_DIR)/onemount-$(VERSION)
	mkdir -p $(TEMP_DIR)/onemount-$(VERSION)
	git ls-files > $(TEMP_DIR)/filelist.txt
	git rev-parse HEAD > $(TEMP_DIR)/.commit
	echo .commit >> $(TEMP_DIR)/filelist.txt
	rsync -a --files-from=$(TEMP_DIR)/filelist.txt . $(TEMP_DIR)/onemount-$(VERSION)
	mv $(TEMP_DIR)/onemount-$(VERSION)/packaging/deb $(TEMP_DIR)/onemount-$(VERSION)/debian
	go mod vendor
	cp -R vendor/ $(TEMP_DIR)/onemount-$(VERSION)
	cd $(TEMP_DIR) && tar -czf ../packages/source/$@ onemount-$(VERSION)
	rm -rf $(TEMP_DIR)/onemount-$(VERSION) vendor/ $(TEMP_DIR)/filelist.txt $(TEMP_DIR)/.commit


# build srpm package used for rpm build with mock
srpm: onemount-$(RPM_FULL_VERSION).src.rpm
onemount-$(RPM_FULL_VERSION).src.rpm: v$(VERSION).tar.gz
	mkdir -p $(RPM_DIR)
	rpmbuild -ts $(SOURCE_DIR)/$<
	cp $$(rpm --eval '%{_topdir}')/SRPMS/$@ $(RPM_DIR)/


# build the rpm for the default mock target
MOCK_CONFIG=$(shell readlink -f /etc/mock/default.cfg | grep -oP '[a-z0-9-]+x86_64')
rpm: onemount-$(RPM_FULL_VERSION).x86_64.rpm
onemount-$(RPM_FULL_VERSION).x86_64.rpm: onemount-$(RPM_FULL_VERSION).src.rpm
	mkdir -p $(RPM_DIR)
	mock -r /etc/mock/$(MOCK_CONFIG).cfg $(RPM_DIR)/$<
	cp /var/lib/mock/$(MOCK_CONFIG)/result/$@ $(RPM_DIR)/


# create a release tarball for debian builds
onemount_$(VERSION).orig.tar.gz: v$(VERSION).tar.gz
	mkdir -p $(DEB_DIR)
	cp $(SOURCE_DIR)/$< $(DEB_DIR)/$@


# create the debian source package for the current version
changes: onemount_$(VERSION)-$(RELEASE)_source.changes
onemount_$(VERSION)-$(RELEASE)_source.changes: onemount_$(VERSION).orig.tar.gz
	mkdir -p $(TEMP_DIR)
	cd $(TEMP_DIR) && tar -xzf ../packages/deb/$< && cd onemount-$(VERSION) && debuild -S -sa -d
	mv $(TEMP_DIR)/onemount_$(VERSION)-$(RELEASE)_source.* $(DEB_DIR)/
	rm -rf $(TEMP_DIR)/onemount-$(VERSION)


# just a helper target to use while building debs
dsc: onemount_$(VERSION)-$(RELEASE).dsc
onemount_$(VERSION)-$(RELEASE).dsc: onemount_$(VERSION).orig.tar.gz
	mkdir -p $(TEMP_DIR)
	cd $(TEMP_DIR) && tar -xzf ../packages/deb/$< && dpkg-source --build onemount-$(VERSION)
	mv $(TEMP_DIR)/onemount_$(VERSION)-$(RELEASE).dsc $(DEB_DIR)/
	rm -rf $(TEMP_DIR)/onemount-$(VERSION)


# create the Ubuntu package using Docker (default)
deb: deb-docker
ubuntu: deb-docker

# create the debian package in a chroot via pbuilder (legacy)
deb-pbuilder: onemount_$(VERSION)-$(RELEASE)_amd64.deb
onemount_$(VERSION)-$(RELEASE)_amd64.deb: onemount_$(VERSION)-$(RELEASE).dsc
	sudo mkdir -p /var/cache/pbuilder/aptcache
	sudo pbuilder --build $(DEB_DIR)/$<
	mkdir -p $(DEB_DIR)
	cp /var/cache/pbuilder/result/$@ $(DEB_DIR)/


clean:
	rm -f *.db *.rpm *.deb *.dsc *.changes *.build* *.upload *.xz filelist.txt .commit
	rm -f *.log *.fa *.gz *.test vgcore.* .auth_tokens.json
	rm -rf util-linux-*/ onemount-*/ vendor/ $(BUILD_DIR)/


# Run all tests
test:
	go test -v -timeout $(TEST_TIMEOUT) ./...

# Run all tests sequentially (no parallel execution)
test-sequential:
	go test -v -p 1 -parallel 1 ./...

# Run unit tests
unit-test:
	go test -v ./... -short

# Run unit tests sequentially (no parallel execution)
unit-test-sequential:
	go test -v -p 1 -parallel 1 ./... -short

# Run integration tests
integration-test:
	go test -v ./internal/testutil/framework/integration_test_env_test.go -timeout $(TEST_TIMEOUT)

# Run system tests
system-test:
	go test -v ./internal/testutil/framework/system_test_env_test.go -timeout $(TEST_TIMEOUT)

# Run comprehensive system tests with real OneDrive account
system-test-real:
	@echo "Running comprehensive system tests with real OneDrive account..."
	./scripts/dev test system --category comprehensive

# Run all system test categories with real OneDrive account
system-test-all:
	@echo "Running all system test categories with real OneDrive account..."
	./scripts/dev test system --category all

# Run performance system tests
system-test-performance:
	@echo "Running performance system tests..."
	./scripts/dev test system --category performance

# Run reliability system tests
system-test-reliability:
	@echo "Running reliability system tests..."
	./scripts/dev test system --category reliability

# Run integration system tests
system-test-integration:
	@echo "Running integration system tests..."
	./scripts/dev test system --category integration

# Run stress system tests
system-test-stress:
	@echo "Running stress system tests..."
	./scripts/dev test system --category stress

# Run system tests directly with Go (alternative to script)
system-test-go:
	go test -v -timeout 30m ./tests/system -run "TestSystemST_.*"

# Run large file system tests (2.5GB+ files)
system-test-large-files:
	@echo "Running large file system tests (requires significant disk space and time)..."
	@echo "WARNING: This test will create files up to 5GB in size and may take 1+ hours to complete"
	@read -p "Continue? [y/N] " -n 1 -r; echo; if [[ ! $$REPLY =~ ^[Yy]$$ ]]; then exit 1; fi
	go test -v -timeout 2h ./tests/system -run "TestSystemST_LARGE_FILES_.*"

# Coverage targets
coverage:
	mkdir -p coverage
	go test -v -coverprofile=coverage/coverage.out ./...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html

coverage-report:
	@echo "Generating comprehensive coverage report..."
	mkdir -p coverage
	go test -v -coverprofile=coverage/coverage.out ./...
	./scripts/dev test coverage

coverage-ci:
	@echo "Running coverage analysis for CI/CD..."
	mkdir -p coverage
	go test -v -coverprofile=coverage/coverage.out ./...
	go tool cover -func=coverage/coverage.out
	./scripts/dev test coverage --ci

coverage-trend:
	@echo "Analyzing coverage trends..."
	./scripts/dev analyze coverage-trends --input coverage/coverage_history.json --output coverage/trends.html

# Docker-based testing targets with enhanced container management
docker-test-build:
	@echo "Building Docker test image (uses pre-built image tags)..."
	./scripts/dev test docker build

docker-test-build-dev:
	@echo "Building Docker development image..."
	./scripts/dev test docker build --dev

docker-test-build-no-cache:
	@echo "Building Docker test image without cache..."
	./scripts/dev test docker build --no-cache

docker-test-build-direct:
	@echo "Building Docker test image with direct Docker (no Compose)..."
	./scripts/dev test docker build --no-compose

docker-test-unit:
	@echo "Running unit tests in Docker (with container reuse)..."
	./scripts/dev --verbose test docker unit

docker-test-unit-fresh:
	@echo "Running unit tests in Docker (fresh container)..."
	./scripts/dev --verbose test docker unit --recreate-container

docker-test-unit-no-reuse:
	@echo "Running unit tests in Docker (no container reuse)..."
	./scripts/dev --verbose test docker unit --no-reuse

docker-test-integration:
	@echo "Running integration tests in Docker..."
	./scripts/dev --verbose test docker integration

docker-test-system:
	@echo "Running system tests in Docker..."
	./scripts/dev --verbose test docker system --timeout 30m

docker-test-all:
	@echo "Running all tests in Docker..."
	./scripts/dev --verbose test docker all

docker-test-coverage:
	@echo "Running coverage analysis in Docker..."
	./scripts/dev --verbose test docker coverage

docker-test-shell:
	@echo "Starting interactive Docker test shell..."
	./scripts/dev test docker shell

docker-test-clean:
	@echo "Cleaning up Docker test resources..."
	./scripts/dev test docker clean

# Development workflow targets
docker-dev-setup:
	@echo "Setting up Docker development environment..."
	@echo "Building tagged image that will be reused for fast testing..."
	./scripts/dev test docker build --dev
	@echo ""
	@echo "✅ Development environment ready!"
	@echo "💡 Images are now tagged and ready for reuse"
	@echo "💡 Use 'make docker-test-unit' for fast testing (5-10 seconds)"
	@echo "💡 Containers will be reused automatically for even faster subsequent runs"

docker-dev-reset:
	@echo "Resetting Docker development environment..."
	./scripts/dev test docker clean
	./scripts/dev test docker build --dev --no-cache
	@echo "✅ Development environment reset with fresh images"
