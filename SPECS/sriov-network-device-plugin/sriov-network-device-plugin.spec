Summary:        Plugin for discovering and advertising networking resources
Name:           sriov-network-device-plugin
Version:        3.7.0
Release:        4%{?dist}
License:        MIT
Vendor:         Microsoft Corporation
Distribution:   Azure Linux
URL:            https://github.com/k8snetworkplumbingwg/sriov-network-device-plugin
Source0:        https://github.com/k8snetworkplumbingwg/%{name}/archive/refs/tags/v%{version}.tar.gz#/%{name}-%{version}.tar.gz
Source1:        %{name}-%{version}-vendor.tar.gz
Patch0:         CVE-2024-45338.patch
Patch1:		CVE-2024-45339.patch
BuildRequires:  golang
Requires:       gawk
Requires:       hwdata

%description
sriov-network-device-plugin is Kubernetes device plugin for discovering and advertising networking
resources in the form of SR-IOV virtual functions and PCI physical functions

%prep
%autosetup -N
tar -xf %{SOURCE1}
%autopatch -p1

%build
go build -mod vendor -o ./build/sriovdp ./cmd/sriovdp/

%install
install -D -m0755 build/sriovdp %{buildroot}%{_bindir}/sriovdp
install -D -m0755 images/entrypoint.sh %{buildroot}%{_bindir}/%{name}-entrypoint.sh
install -D -m0755 images/ddptool-1.0.1.12.tar.gz %{buildroot}%{_datadir}/%{name}/ddptool-1.0.1.12.tar.gz

%files
%license LICENSE
%doc README.md
%{_bindir}/sriovdp
%{_bindir}/%{name}-entrypoint.sh
%{_datadir}/%{name}/ddptool-1.0.1.12.tar.gz

%changelog
* Fri Mar 21 2025 Anuj Mittal <anuj.mittal@intel.com> - 3.7.0-4
- Bump Release to rebuild

* Fri Jan 31 2025 Kavya Sree Kaitepalli <kkaitepalli@microsoft.com> - 3.7.0-3
- Patch CVE-2024-45339

* Tue Dec 31 2024 Rohit Rawat <rohitrawat@microsoft.com> - 3.7.0-2
- Patch CVE-2024-45338

* Thu Jun 06 2024 CBL-Mariner Servicing Account <cblmargh@microsoft.com> - 3.7.0-1
- Auto-upgrade to 3.7.0 - address CVE-2022-1996

* Mon Oct 16 2023 CBL-Mariner Servicing Account <cblmargh@microsoft.com> - 3.5.1-3
- Bump release to rebuild with go 1.20.10

* Tue Oct 10 2023 Dan Streetman <ddstreet@ieee.org> - 3.5.1-2
- Bump release to rebuild with updated version of Go.

* Thu Sep 28 2023 Aditya Dubey <adityadubey@microsoft.com> - 3.5.1-1
- Upgrade to v3.5.1

* Mon Aug 07 2023 CBL-Mariner Servicing Account <cblmargh@microsoft.com> - 3.4.0-12
- Bump release to rebuild with go 1.19.12

* Thu Jul 13 2023 CBL-Mariner Servicing Account <cblmargh@microsoft.com> - 3.4.0-11
- Bump release to rebuild with go 1.19.11

* Thu Jun 15 2023 CBL-Mariner Servicing Account <cblmargh@microsoft.com> - 3.4.0-10
- Bump release to rebuild with go 1.19.10

* Wed Apr 05 2023 CBL-Mariner Servicing Account <cblmargh@microsoft.com> - 3.4.0-9
- Bump release to rebuild with go 1.19.8

* Tue Mar 28 2023 CBL-Mariner Servicing Account <cblmargh@microsoft.com> - 3.4.0-8
- Bump release to rebuild with go 1.19.7

* Wed Mar 15 2023 CBL-Mariner Servicing Account <cblmargh@microsoft.com> - 3.4.0-7
- Bump release to rebuild with go 1.19.6

* Fri Feb 03 2023 CBL-Mariner Servicing Account <cblmargh@microsoft.com> - 3.4.0-6
- Bump release to rebuild with go 1.19.5

* Wed Jan 18 2023 CBL-Mariner Servicing Account <cblmargh@microsoft.com> - 3.4.0-5
- Bump release to rebuild with go 1.19.4

* Fri Dec 16 2022 Daniel McIlvaney <damcilva@microsoft.com> - 3.4.0-4
- Bump release to rebuild with go 1.18.8 with patch for CVE-2022-41717

* Tue Dec 06 2022 Aditya Dubey <adityadubey@microsoft.com> - 3.4.0-3
- Adding in the hwdata and gawk dependencies

* Tue Nov 01 2022 Olivia Crain <oliviacrain@microsoft.com> - 3.4.0-2
- Bump release to rebuild with go 1.18.8

* Fri Sep 23 2022 Aditya Dubey <adityadubey@microsoft.com> - 3.4.0-1
- Original version for CBL-Mariner
- License Verified
