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
	mkdir -p $(HOME)/.local/bin/
	mkdir -p $(HOME)/.local/share/icons/onemount/
	mkdir -p $(HOME)/.local/share/applications/
	mkdir -p $(HOME)/.config/systemd/user/
	mkdir -p $(HOME)/.local/share/man/man1/
	cp $(OUTPUT_DIR)/onemount $(HOME)/.local/bin/
	cp $(OUTPUT_DIR)/onemount-launcher $(HOME)/.local/bin/
	cp assets/icons/OneMount-Logo.svg $(HOME)/.local/share/icons/onemount/
	cp assets/icons/OneMount-Logo-Icon.svg $(HOME)/.local/share/icons/onemount/
	cp assets/icons/OneMount.png $(HOME)/.local/share/icons/onemount/onemount.png
	cp assets/icons/OneMount-Logo-128.png $(HOME)/.local/share/icons/onemount/onemount-128.png
	sed -e 's|@BIN_PATH@|$(HOME)/.local/bin|g' \
		-e 's|@ICON_PATH@|$(HOME)/.local/share/icons/onemount|g' \
		deployments/desktop/onemount-launcher.desktop.template > $(HOME)/.local/share/applications/onemount-launcher.desktop
	sed -e 's|@BIN_PATH@|$(HOME)/.local/bin|g' \
		-e 's|@AFTER@||g' \
		-e 's|@USER@||g' \
		-e 's|@GROUP@||g' \
		-e 's|@WANTED_BY@|default.target|g' \
		deployments/systemd/onemount@.service.template > $(HOME)/.config/systemd/user/onemount@.service
	systemctl --user daemon-reload 2>/dev/null || true
	gzip -c docs/man/onemount.1 > $(HOME)/.local/share/man/man1/onemount.1.gz
	mandb --user-db --quiet 2>/dev/null || true


install-system: onemount onemount-launcher
	sudo mkdir -p /usr/local/bin/
	sudo mkdir -p /usr/local/share/icons/onemount/
	sudo mkdir -p /usr/local/share/applications/
	sudo mkdir -p /usr/local/lib/systemd/system/
	sudo mkdir -p /usr/local/share/man/man1/
	sudo cp $(OUTPUT_DIR)/onemount /usr/local/bin/
	sudo cp $(OUTPUT_DIR)/onemount-launcher /usr/local/bin/
	sudo cp assets/icons/OneMount-Logo.svg /usr/local/share/icons/onemount/
	sudo cp assets/icons/OneMount-Logo-Icon.svg /usr/local/share/icons/onemount/
	sudo cp assets/icons/OneMount.png /usr/local/share/icons/onemount/onemount.png
	sudo cp assets/icons/OneMount-Logo-128.png /usr/local/share/icons/onemount/onemount-128.png
	sudo sed -e 's|@BIN_PATH@|/usr/local/bin|g' \
		-e 's|@ICON_PATH@|/usr/local/share/icons/onemount|g' \
		deployments/desktop/onemount-launcher.desktop.template > /usr/local/share/applications/onemount-launcher.desktop
	sudo sed -e 's|@BIN_PATH@|/usr/local/bin|g' \
		-e 's|@AFTER@|\nAfter=network.target|g' \
		-e 's|@USER@|\nUser=%i|g' \
		-e 's|@GROUP@|\nGroup=%i|g' \
		-e 's|@WANTED_BY@|multi-user.target|g' \
		deployments/systemd/onemount@.service.template > /usr/local/lib/systemd/system/onemount@.service
	sudo systemctl daemon-reload
	sudo gzip -c docs/man/onemount.1 > /usr/local/share/man/man1/onemount.1.gz
	sudo mandb


uninstall:
	rm -f \
		$(HOME)/.local/bin/onemount \
		$(HOME)/.local/bin/onemount-launcher \
		$(HOME)/.config/systemd/user/onemount@.service \
		$(HOME)/.local/share/applications/onemount-launcher.desktop \
		$(HOME)/.local/share/man/man1/onemount.1.gz
	rm -rf $(HOME)/.local/share/icons/onemount
	systemctl --user daemon-reload 2>/dev/null || true
	mandb --user-db --quiet 2>/dev/null || true


uninstall-system:
	sudo rm -f \
		/usr/local/bin/onemount \
		/usr/local/bin/onemount-launcher \
		/usr/local/lib/systemd/system/onemount@.service \
		/usr/local/share/applications/onemount-launcher.desktop \
		/usr/local/share/man/man1/onemount.1.gz
	sudo rm -rf /usr/local/share/icons/onemount
	sudo systemctl daemon-reload
	sudo mandb


# Validate packaging requirements
validate-packaging:
	@echo "Validating packaging requirements..."
	@test -f docs/man/onemount.1 || (echo "Error: docs/man/onemount.1 not found" && exit 1)
	@test -f assets/icons/OneMount.png || (echo "Error: assets/icons/OneMount.png not found" && exit 1)
	@test -f assets/icons/OneMount-Logo-128.png || (echo "Error: assets/icons/OneMount-Logo-128.png not found" && exit 1)
	@test -f assets/icons/OneMount-Logo.svg || (echo "Error: assets/icons/OneMount-Logo.svg not found" && exit 1)
	@test -f deployments/desktop/onemount-launcher.desktop.template || (echo "Error: desktop file template not found" && exit 1)
	@test -f deployments/systemd/onemount@.service.template || (echo "Error: systemd service file template not found" && exit 1)
	@test -f scripts/cgo-helper.sh || (echo "Error: cgo-helper.sh script not found" && exit 1)
	@echo "All packaging requirements validated successfully"

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
