Name:          onemount
Version:       0.1.0rc1
Release:       1%{?dist}
Summary:       Linux access to OneDrive made simple.

License:       GPL-3.0-or-later
URL:           https://github.com/auriora/onemount
Source0:       https://github.com/auriora/onemount/archive/refs/tags/v%{version}.tar.gz

%if 0%{?suse_version}
BuildRequires: go >= 1.23
%else
BuildRequires: golang >= 1.23.0
%endif
BuildRequires: git
BuildRequires: gcc
BuildRequires: pkg-config
BuildRequires: webkit2gtk3-devel
BuildRequires: gzip
Requires: fuse3
Requires: webkit2gtk3
Suggests: python3-nemo

%description
OneMount is a network filesystem that gives your computer direct access to your
files on Microsoft OneDrive. This is not a sync client. Instead of syncing
files, OneMount performs an on-demand download of files when your computer
attempts to use them. OneMount allows you to use files on OneDrive as if they
were files on your local computer.

%prep
%autosetup

%build
bash scripts/cgo-helper.sh
mkdir -p build/binaries
if rpm -q pango | grep -q 1.42; then
  BUILD_TAGS=-tags=pango_1_42,gtk_3_22
fi
go build -v -mod=vendor $BUILD_TAGS \
  -o build/binaries/onemount \
  -ldflags="-X github.com/auriora/onemount/cmd/common.commit=$(cat .commit)" \
  ./cmd/onemount
go build -v -mod=vendor $BUILD_TAGS \
  -o build/binaries/onemount-launcher \
  -ldflags="-X github.com/auriora/onemount/cmd/common.commit=$(cat .commit)" \
  ./cmd/onemount-launcher
gzip docs/man/onemount.1

%install
rm -rf $RPM_BUILD_ROOT
# Use centralized installation manifest (standalone version)
python3 scripts/install-manifest-standalone.py --target rpm --action install | bash

# fix for el8 build in mock
%define _empty_manifest_terminate_build 0
%files
# Use centralized installation manifest for files list (standalone version)
%(python3 scripts/install-manifest-standalone.py --target rpm --action files)

%changelog
* Sun Jun 01 2025 Bruce Cherrington <aurioraproject@gmail.com> - 0.1.0rc1
- Beta release with enhanced conflict resolution and synchronization
- Comprehensive offline-to-online transition support
- Improved error handling and recovery mechanisms
- Full test coverage for filesystem operations
