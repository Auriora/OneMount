.PHONY: all, test, test-init, test-python, srpm, rpm, dsc, changes, deb, clean, install, uninstall, update-imports

# autocalculate software/package versions
VERSION := $(shell grep Version scripts/onedriver.spec | sed 's/Version: *//g')
RELEASE := $(shell grep -oP "Release: *[0-9]+" scripts/onedriver.spec | sed 's/Release: *//g')
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

all: onedriver onedriver-launcher


onedriver: $(shell find internal/fs/ -type f) cmd/onedriver/main.go
	bash scripts/cgo-helper.sh
	mkdir -p $(OUTPUT_DIR)
	$(CGO_CFLAGS) go build -v \
		-o $(OUTPUT_DIR)/onedriver \
		-ldflags="-X github.com/bcherrington/onedriver/cmd/common.commit=$(shell git rev-parse HEAD)" \
		./cmd/onedriver


onedriver-headless: $(shell find internal/fs/ cmd/common/ -type f) cmd/onedriver/main.go
	CGO_ENABLED=0 go build -v \
		-o $(OUTPUT_DIR)/onedriver/onedriver-headless \
		-ldflags="-X github.com/bcherrington/onedriver/cmd/common.commit=$(shell git rev-parse HEAD)" \
		./cmd/onedriver


onedriver-launcher: $(shell find internal/ui/ cmd/common/ -type f) cmd/onedriver-launcher/main.go
	$(CGO_CFLAGS) go build -v \
		-o $(OUTPUT_DIR)/onedriver-launcher \
		-ldflags="-X github.com/bcherrington/onedriver/cmd/common.commit=$(shell git rev-parse HEAD)" \
		./cmd/onedriver-launcher


install: onedriver onedriver-launcher
	cp $(OUTPUT_DIR)/onedriver /usr/bin/
	cp $(OUTPUT_DIR)/onedriver-launcher /usr/bin/
	mkdir -p /usr/share/icons/onedriver/
	cp configs/resources/onedriver.svg /usr/share/icons/onedriver/
	cp configs/resources/onedriver.png /usr/share/icons/onedriver/
	cp configs/resources/onedriver-128.png /usr/share/icons/onedriver/
	cp configs/resources/onedriver-launcher.desktop /usr/share/applications/
	cp configs/resources/onedriver@.service /etc/systemd/user/
	gzip -c configs/resources/onedriver.1 > /usr/share/man/man1/onedriver.1.gz
	mandb


uninstall:
	rm -f \
		/usr/bin/onedriver \
		/usr/bin/onedriver-launcher \
		/etc/systemd/user/onedriver@.service \
		/usr/share/applications/onedriver-launcher.desktop \
		/usr/share/man/man1/onedriver.1.gz
	rm -rf /usr/share/icons/onedriver
	mandb


# used to create release tarball for rpmbuild
v$(VERSION).tar.gz: $(shell git ls-files)
	rm -rf onedriver-$(VERSION)
	mkdir -p onedriver-$(VERSION)
	git ls-files > filelist.txt
	git rev-parse HEAD > .commit
	echo .commit >> filelist.txt
	rsync -a --files-from=filelist.txt . onedriver-$(VERSION)
	mv onedriver-$(VERSION)/pkg/debian onedriver-$(VERSION)
	go mod vendor
	cp -R vendor/ onedriver-$(VERSION)
	tar -czf $@ onedriver-$(VERSION)


# build srpm package used for rpm build with mock
srpm: onedriver-$(RPM_FULL_VERSION).src.rpm 
onedriver-$(RPM_FULL_VERSION).src.rpm: v$(VERSION).tar.gz
	rpmbuild -ts $<
	cp $$(rpm --eval '%{_topdir}')/SRPMS/$@ .


# build the rpm for the default mock target
MOCK_CONFIG=$(shell readlink -f /etc/mock/default.cfg | grep -oP '[a-z0-9-]+x86_64')
rpm: onedriver-$(RPM_FULL_VERSION).x86_64.rpm
onedriver-$(RPM_FULL_VERSION).x86_64.rpm: onedriver-$(RPM_FULL_VERSION).src.rpm
	mock -r /etc/mock/$(MOCK_CONFIG).cfg $<
	cp /var/lib/mock/$(MOCK_CONFIG)/result/$@ .


# create a release tarball for debian builds
onedriver_$(VERSION).orig.tar.gz: v$(VERSION).tar.gz
	cp $< $@


# create the debian source package for the current version
changes: onedriver_$(VERSION)-$(RELEASE)_source.changes
onedriver_$(VERSION)-$(RELEASE)_source.changes: onedriver_$(VERSION).orig.tar.gz
	cd onedriver-$(VERSION) && debuild -S -sa -d


# just a helper target to use while building debs
dsc: onedriver_$(VERSION)-$(RELEASE).dsc
onedriver_$(VERSION)-$(RELEASE).dsc: onedriver_$(VERSION).orig.tar.gz
	dpkg-source --build onedriver-$(VERSION)


# create the debian package in a chroot via pbuilder
deb: onedriver_$(VERSION)-$(RELEASE)_amd64.deb
onedriver_$(VERSION)-$(RELEASE)_amd64.deb: onedriver_$(VERSION)-$(RELEASE).dsc
	sudo mkdir -p /var/cache/pbuilder/aptcache
	sudo pbuilder --build $<
	cp /var/cache/pbuilder/result/$@ .

# setup tests for the first time on a new computer
test-init: onedriver
	go install github.com/rakyll/gotest@latest
	pip install pytest pytest-mock

# Run Python tests for nemo-onedriver.py
# PYTHONPATH is set to include the current directory and nemo-onedriver/src
# to help pytest find modules and prevent import errors
# The test file has been modified to mock D-Bus and GLib to prevent hanging during collection
test-python:
	PYTHONPATH=.:internal/nemo/src pytest -xvs internal/nemo/tests/test_nemo_onedriver.py

# For offline tests, the test binary is built online, then network access is
# disabled and tests are run. sudo is required - otherwise we don't have
# permission to deny network access to onedriver during the test.
test: onedriver onedriver-launcher test-python
	CGO_ENABLED=0 gotest -v -parallel=1 -count=1 ./internal/ui/...
	$(CGO_CFLAGS) $(GORACE) gotest -race -v -parallel=1 -count=1 ./internal/fs/...
	$(CGO_CFLAGS) $(GORACE) gotest -race -v -parallel=1 -count=1 ./cmd/...


clean:
	fusermount3 -uz mount/ || true
	rm -f *.db *.rpm *.deb *.dsc *.changes *.build* *.upload *.xz filelist.txt .commit
	rm -f *.log *.fa *.gz *.test vgcore.* onedriver onedriver-headless onedriver-launcher .auth_tokens.json
	rm -rf util-linux-*/ onedriver-*/ vendor/ build/
