.PHONY: all, test, test-init, test-python, unit-test, integration-test, system-test, srpm, rpm, dsc, changes, deb, clean, install, install-system, uninstall, uninstall-system, validate-packaging, update-imports

# auto-calculate software/package versions
VERSION := $(shell grep Version packaging/rpm/onemount.spec | sed 's/Version: *//g')
RELEASE := $(shell grep -oP "Release: *[0-9]+" packaging/rpm/onemount.spec | sed 's/Release: *//g')
DIST := $(shell rpm --eval "%{?dist}" 2> /dev/null || echo 1)
RPM_FULL_VERSION = $(VERSION)-$(RELEASE)$(DIST)

# -Wno-deprecated-declarations is for gotk3, which uses deprecated methods for older
# glib compatibility: https://github.com/gotk3/gotk3/issues/762#issuecomment-919035313
CGO_CFLAGS := CGO_CFLAGS=-Wno-deprecated-declarations

OUTPUT_DIR := build

# test-specific variables
GORACE := GORACE="log_path=fusefs_tests.race strip_path_prefix=1"
TEST_TIMEOUT := 5m

all: onemount onemount-launcher


onemount: $(shell find internal/fs/ pkg/ -type f) cmd/onemount/main.go
	bash scripts/cgo-helper.sh
	mkdir -p $(OUTPUT_DIR)
	$(CGO_CFLAGS) go build -v \
		-o $(OUTPUT_DIR)/onemount \
		-ldflags="-X github.com/auriora/onemount/cmd/common.commit=$(shell git rev-parse HEAD)" \
		./cmd/onemount


onemount-headless: $(shell find internal/fs/ cmd/common/ pkg/ -type f) cmd/onemount/main.go
	mkdir -p $(OUTPUT_DIR)
	CGO_ENABLED=0 go build -v \
		-o $(OUTPUT_DIR)/onemount-headless \
		-ldflags="-X github.com/auriora/onemount/cmd/common.commit=$(shell git rev-parse HEAD)" \
		./cmd/onemount


onemount-launcher: $(shell find internal/ui/ cmd/common/ pkg/ -type f) cmd/onemount-launcher/main.go
	$(CGO_CFLAGS) go build -v \
		-o $(OUTPUT_DIR)/onemount-launcher \
		-ldflags="-X github.com/auriora/onemount/cmd/common.commit=$(shell git rev-parse HEAD)" \
		./cmd/onemount-launcher


install: onemount onemount-launcher
	@python3 scripts/install-manifest.py --target makefile --type user --action install | bash


install-system: onemount onemount-launcher
	@python3 scripts/install-manifest.py --target makefile --type system --action install | bash


uninstall:
	@python3 scripts/install-manifest.py --target makefile --type user --action uninstall | bash


uninstall-system:
	@python3 scripts/install-manifest.py --target makefile --type system --action uninstall | bash


# Validate packaging requirements
validate-packaging:
	@python3 scripts/install-manifest.py --target makefile --action validate | bash
	@test -f scripts/cgo-helper.sh || (echo "Error: cgo-helper.sh script not found" && exit 1)

# used to create release tarball for rpmbuild
v$(VERSION).tar.gz: validate-packaging $(shell git ls-files)
	rm -rf onemount-$(VERSION)
	mkdir -p onemount-$(VERSION)
	git ls-files > filelist.txt
	git rev-parse HEAD > .commit
	echo .commit >> filelist.txt
	rsync -a --files-from=filelist.txt . onemount-$(VERSION)
	mv onemount-$(VERSION)/packaging/deb onemount-$(VERSION)/debian
	go mod vendor
	cp -R vendor/ onemount-$(VERSION)
	tar -czf $@ onemount-$(VERSION)


# build srpm package used for rpm build with mock
srpm: onemount-$(RPM_FULL_VERSION).src.rpm
onemount-$(RPM_FULL_VERSION).src.rpm: v$(VERSION).tar.gz
	rpmbuild -ts $<
	cp $$(rpm --eval '%{_topdir}')/SRPMS/$@ .


# build the rpm for the default mock target
MOCK_CONFIG=$(shell readlink -f /etc/mock/default.cfg | grep -oP '[a-z0-9-]+x86_64')
rpm: onemount-$(RPM_FULL_VERSION).x86_64.rpm
onemount-$(RPM_FULL_VERSION).x86_64.rpm: onemount-$(RPM_FULL_VERSION).src.rpm
	mock -r /etc/mock/$(MOCK_CONFIG).cfg $<
	cp /var/lib/mock/$(MOCK_CONFIG)/result/$@ .


# create a release tarball for debian builds
onemount_$(VERSION).orig.tar.gz: v$(VERSION).tar.gz
	cp $< $@


# create the debian source package for the current version
changes: onemount_$(VERSION)-$(RELEASE)_source.changes
onemount_$(VERSION)-$(RELEASE)_source.changes: onemount_$(VERSION).orig.tar.gz
	cd onemount-$(VERSION) && debuild -S -sa -d


# just a helper target to use while building debs
dsc: onemount_$(VERSION)-$(RELEASE).dsc
onemount_$(VERSION)-$(RELEASE).dsc: onemount_$(VERSION).orig.tar.gz
	dpkg-source --build onemount-$(VERSION)


# create the debian package in a chroot via pbuilder
deb: onemount_$(VERSION)-$(RELEASE)_amd64.deb
onemount_$(VERSION)-$(RELEASE)_amd64.deb: onemount_$(VERSION)-$(RELEASE).dsc
	sudo mkdir -p /var/cache/pbuilder/aptcache
	sudo pbuilder --build $<
	cp /var/cache/pbuilder/result/$@ .


clean:
	rm -f *.db *.rpm *.deb *.dsc *.changes *.build* *.upload *.xz filelist.txt .commit
	rm -f *.log *.fa *.gz *.test vgcore.* .auth_tokens.json
	rm -f $(OUTPUT_DIR)/onemount $(OUTPUT_DIR)/onemount-headless $(OUTPUT_DIR)/onemount-launcher
	rm -rf util-linux-*/ onemount-*/ vendor/ $(OUTPUT_DIR)/


# Run all tests
test:
	go test -v ./...

# Run unit tests
unit-test:
	go test -v ./... -short

# Run integration tests
integration-test:
	go test -v ./pkg/testutil/integration_test_env_test.go -timeout $(TEST_TIMEOUT)

# Run system tests
system-test:
	go test -v ./pkg/testutil/system_test_env_test.go -timeout $(TEST_TIMEOUT)

# Run comprehensive system tests with real OneDrive account
system-test-real:
	@echo "Running comprehensive system tests with real OneDrive account..."
	./scripts/run-system-tests.sh --comprehensive

# Run all system test categories with real OneDrive account
system-test-all:
	@echo "Running all system test categories with real OneDrive account..."
	./scripts/run-system-tests.sh --all

# Run performance system tests
system-test-performance:
	@echo "Running performance system tests..."
	./scripts/run-system-tests.sh --performance

# Run reliability system tests
system-test-reliability:
	@echo "Running reliability system tests..."
	./scripts/run-system-tests.sh --reliability

# Run integration system tests
system-test-integration:
	@echo "Running integration system tests..."
	./scripts/run-system-tests.sh --integration

# Run stress system tests
system-test-stress:
	@echo "Running stress system tests..."
	./scripts/run-system-tests.sh --stress

# Run system tests directly with Go (alternative to script)
system-test-go:
	go test -v -timeout 30m ./tests/system -run "TestSystemST_.*"

# Coverage targets
coverage:
	mkdir -p coverage
	go test -v -coverprofile=coverage/coverage.out ./...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html

coverage-report:
	@echo "Generating comprehensive coverage report..."
	mkdir -p coverage
	go test -v -coverprofile=coverage/coverage.out ./...
	bash scripts/coverage-report.sh

coverage-ci:
	@echo "Running coverage analysis for CI/CD..."
	mkdir -p coverage
	go test -v -coverprofile=coverage/coverage.out ./...
	go tool cover -func=coverage/coverage.out
	bash scripts/coverage-report.sh --ci

coverage-trend:
	@echo "Analyzing coverage trends..."
	python3 scripts/coverage-trend-analysis.py --input coverage/coverage_history.json --output coverage/trends.html
