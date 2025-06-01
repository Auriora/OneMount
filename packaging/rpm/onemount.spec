Name:          onemount
Version:       0.14.1
Release:       1%{?dist}
Summary:       Linux access to OneDrive made simple.

License:       GPL-3.0-or-later
URL:           https://github.com/auriora/onemount
Source0:       https://github.com/auriora/onemount/archive/refs/tags/v%{version}.tar.gz

%if 0%{?suse_version}
BuildRequires: go >= 1.21
%else
BuildRequires: golang >= 1.21.0
%endif
BuildRequires: git
BuildRequires: gcc
BuildRequires: pkg-config
BuildRequires: webkit2gtk3-devel
BuildRequires: gzip
Requires: fuse3
Requires: webkit2gtk3

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
if rpm -q pango | grep -q 1.42; then
  BUILD_TAGS=-tags=pango_1_42,gtk_3_22
fi
go build -v -mod=vendor $BUILD_TAGS \
  -ldflags="-X github.com/auriora/onemount/cmd/common.commit=$(cat .commit)" \
  ./cmd/onemount
go build -v -mod=vendor $BUILD_TAGS \
  -ldflags="-X github.com/auriora/onemount/cmd/common.commit=$(cat .commit)" \
  ./cmd/onemount-launcher
gzip docs/man/onemount.1

%install
rm -rf $RPM_BUILD_ROOT
mkdir -p %{buildroot}/%{_bindir}
mkdir -p %{buildroot}/usr/share/icons/%{name}
mkdir -p %{buildroot}/usr/share/applications
mkdir -p %{buildroot}/usr/lib/systemd/system
mkdir -p %{buildroot}/usr/share/man/man1
cp %{name} %{buildroot}/%{_bindir}
cp %{name}-launcher %{buildroot}/%{_bindir}
cp assets/icons/OneMount.png %{buildroot}/usr/share/icons/%{name}/%{name}.png
cp assets/icons/OneMount-Logo-128.png %{buildroot}/usr/share/icons/%{name}/%{name}-128.png
cp assets/icons/OneMount-Logo.svg %{buildroot}/usr/share/icons/%{name}
cp deployments/desktop/%{name}-launcher-system.desktop %{buildroot}/usr/share/applications/%{name}-launcher.desktop
cp deployments/systemd/%{name}@-system.service %{buildroot}/usr/lib/systemd/system/%{name}@.service
cp docs/man/%{name}.1.gz %{buildroot}/usr/share/man/man1

# fix for el8 build in mock
%define _empty_manifest_terminate_build 0
%files
%defattr(-,root,root,-)
%attr(755, root, root) %{_bindir}/%{name}
%attr(755, root, root) %{_bindir}/%{name}-launcher
%dir /usr/share/icons/%{name}
%attr(644, root, root) /usr/share/icons/%{name}/%{name}.png
%attr(644, root, root) /usr/share/icons/%{name}/%{name}-128.png
%attr(644, root, root) /usr/share/icons/%{name}/OneMount-Logo.svg
%attr(644, root, root) /usr/share/applications/%{name}-launcher.desktop
%attr(644, root, root) /usr/lib/systemd/system/%{name}@.service
%doc
%attr(644, root, root) /usr/share/man/man1/%{name}.1.gz

%changelog
* Sun Jun 01 2025 Bruce Cherrington <aurioraproject@gmail.com> - 0.14.1-1
- Beta release with enhanced conflict resolution and synchronization
- Comprehensive offline-to-online transition support
- Improved error handling and recovery mechanisms
- Full test coverage for filesystem operations
