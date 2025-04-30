.PHONY: all, test, test-init, test-python, srpm, rpm, dsc, changes, deb, clean, install, uninstall, update-imports

# autocalculate software/package versions
VERSION := $(shell grep Version scripts/onemount.spec | sed 's/Version: *//g')
RELEASE := $(shell grep -oP "Release: *[0-9]+" scripts/onemount.spec | sed 's/Release: *//g')
DIST := $(shell rpm --eval "%{?dist}" 2> /dev/null || echo 1)
RPM_FULL_VERSION = $(VERSION)-$(RELEASE)$(DIST)

# -Wno-deprecated-declarations is for gotk3, which uses deprecated methods for older
# glib compatibility: https://github.com/gotk3/gotk3/issues/762#issuecomment-919035313
CGO_CFLAGS := CGO_CFLAGS=-Wno-deprecated-declarations

# Add this near the top with other variables
OUTPUT_DIR := build

# test-specific variables
TEST_UID := $(shell whoami)
GORACE := GORACE="log_path=fusefs_tests.race strip_path_prefix=1"

all: onemount onemount-launcher


onemount: $(shell find internal/fs/ -type f) cmd/onemount/main.go
	bash scripts/cgo-helper.sh
	mkdir -p $(OUTPUT_DIR)
	$(CGO_CFLAGS) go build -v \
		-o $(OUTPUT_DIR)/onemount \
		-ldflags="-X github.com/bcherrington/onemount/cmd/common.commit=$(shell git rev-parse HEAD)" \
		./cmd/onemount


onemount-headless: $(shell find internal/fs/ cmd/common/ -type f) cmd/onemount/main.go
	CGO_ENABLED=0 go build -v \
		-o $(OUTPUT_DIR)/onemount/onemount-headless \
		-ldflags="-X github.com/bcherrington/onemount/cmd/common.commit=$(shell git rev-parse HEAD)" \
		./cmd/onemount


onemount-launcher: $(shell find internal/ui/ cmd/common/ -type f) cmd/onemount-launcher/main.go
	$(CGO_CFLAGS) go build -v \
		-o $(OUTPUT_DIR)/onemount-launcher \
		-ldflags="-X github.com/bcherrington/onemount/cmd/common.commit=$(shell git rev-parse HEAD)" \
		./cmd/onemount-launcher


install: onemount onemount-launcher
	cp $(OUTPUT_DIR)/onemount /usr/bin/
	cp $(OUTPUT_DIR)/onemount-launcher /usr/bin/
	mkdir -p /usr/share/icons/onemount/
	cp configs/resources/OneMount-Logo.svg /usr/share/icons/onemount/
	cp configs/resources/onemount.png /usr/share/icons/onemount/
	cp configs/resources/onemount-128.png /usr/share/icons/onemount/
	cp configs/resources/onemount-launcher.desktop /usr/share/applications/
	cp configs/resources/onemount@.service /etc/systemd/user/
	gzip -c configs/resources/onemount.1 > /usr/share/man/man1/onemount.1.gz
	mandb


uninstall:
	rm -f \
		/usr/bin/onemount \
		/usr/bin/onemount-launcher \
		/etc/systemd/user/onemount@.service \
		/usr/share/applications/onemount-launcher.desktop \
		/usr/share/man/man1/onemount.1.gz
	rm -rf /usr/share/icons/onemount
	mandb


# used to create release tarball for rpmbuild
v$(VERSION).tar.gz: $(shell git ls-files)
	rm -rf onemount-$(VERSION)
	mkdir -p onemount-$(VERSION)
	git ls-files > filelist.txt
	git rev-parse HEAD > .commit
	echo .commit >> filelist.txt
	rsync -a --files-from=filelist.txt . onemount-$(VERSION)
	mv onemount-$(VERSION)/pkg/debian onemount-$(VERSION)
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

# setup tests for the first time on a new computer
test-init: onemount
	go install github.com/rakyll/gotest@latest
	pip install pytest pytest-mock

# Run Python tests for nemo-onemount.py
# PYTHONPATH is set to include the current directory and nemo-onemount/src
# to help pytest find modules and prevent import errors
# The test file has been modified to mock D-Bus and GLib to prevent hanging during collection
test-python:
	PYTHONPATH=.:internal/nemo/src pytest -xvs internal/nemo/tests/test_nemo_onemount.py

# For offline tests, the test binary is built online, then network access is
# disabled and tests are run. sudo is required - otherwise we don't have
# permission to deny network access to onemount during the test.
test: onemount onemount-launcher test-python
	CGO_ENABLED=0 gotest -v -parallel=1 -count=1 ./internal/ui/...
	$(CGO_CFLAGS) $(GORACE) gotest -race -v -parallel=1 -count=1 ./internal/fs/...
	$(CGO_CFLAGS) $(GORACE) gotest -race -v -parallel=1 -count=1 ./cmd/...


clean:
	fusermount3 -uz mount/ || true
	rm -f *.db *.rpm *.deb *.dsc *.changes *.build* *.upload *.xz filelist.txt .commit
	rm -f *.log *.fa *.gz *.test vgcore.* onemount onemount-headless onemount-launcher .auth_tokens.json
	rm -rf util-linux-*/ onemount-*/ vendor/ build/
