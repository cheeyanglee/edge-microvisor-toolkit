Summary:        A Python XML Schema validator and decoder
Name:           python-xmlschema
Version:        3.2.1
Release:        7%{?dist}
License:        MIT
Vendor:         Intel Corporation
Distribution:   Edge Microvisor Toolkit
URL:            https://github.com/brunato/xmlschema
Source0:        https://github.com/sissaschool/xmlschema/archive/refs/tags/v%{version}.tar.gz#/%{name}-%{version}.tar.gz
%global pypi_name xmlschema
%global _description %{expand:
The xmlschema library is an implementation of XML Schema for Python.
This library arises from the needs of a solid Python layer for processing XML
Schema based files for MaX (Materials design at the Exascale) European project.
A significant problem is the encoding and the decoding of the XML data files
produced by different simulation software. Another important requirement is
the XML data validation, in order to put the produced data under control.
The lack of a suitable alternative for Python in the schema-based decoding
of XML data has led to build this library. Obviously this library can be
useful for other cases related to XML Schema based processing, not only for
the original scope.}
BuildRequires:  pyproject-rpm-macros
BuildRequires:  python3-devel
BuildRequires:  python3-pip
BuildRequires:  python3-wheel
BuildArch:      noarch

%description %{_description}

%package -n     python3-%{pypi_name}
Summary:        %{summary}
Requires:       python3-elementpath
%{?python_provide:%python_provide python3-%{pypi_name}}

%description -n python3-%{pypi_name}  %{_description}

%prep
%autosetup -p1 -n %{pypi_name}-%{version}
sed -i 's/~=/>=/' setup.py tox.ini  # https://bugzilla.redhat.com/show_bug.cgi?id=1758141
sed -i 's/==/>=/' tox.ini  # too strict test deps
sed -i '/memory_profiler/d' tox.ini # optional test dep, not packaged in Fedora, not worth testing
%py3_shebang_fix %{pypi_name}

%generate_buildrequires
%pyproject_buildrequires -t

%build
%pyproject_wheel

%install
%pyproject_install

%check
%tox

%files -n python3-%{pypi_name}
%license LICENSE
%doc README.rst
%{python3_sitelib}/%{pypi_name}/
%{python3_sitelib}/%{pypi_name}-%{version}.dist-info/
%{_bindir}/xmlschema-json2xml
%{_bindir}/xmlschema-validate
%{_bindir}/xmlschema-xml2json

%changelog
* Mon Dec 30 2024 Naveen Saini <naveen.kumar.saini@intel.com> - 3.2.1-7
- Update Source URL.

* Thu Dec 19 2024 Mun Chun Yep <mun.chun.yep@intel.com> - 3.2.1-6
- Add vendor and distribution tag.

* Thu Nov 14 2024 Mun Chun Yep <mun.chun.yep@intel.com> - 3.2.1-5
- Fix elementpath not included in image.

* Wed Oct 16 2024 Mun Chun Yep <mun.chun.yep@intel.com> - 3.2.1-4
- Include elementpath as runtime dependency.

* Mon Sep 09 2024 Mun Chun Yep <mun.chun.yep@intel.com> - 3.2.1-3
- Initial Edge Microvisor Toolkit import from Fedora 41 (license: MIT). License verified.

* Fri Jan 26 2024 Fedora Release Engineering <releng@fedoraproject.org> - 2.4.0-3
- Rebuilt for https://fedoraproject.org/wiki/Fedora_40_Mass_Rebuild

* Mon Jan 22 2024 Fedora Release Engineering <releng@fedoraproject.org> - 2.4.0-2
- Rebuilt for https://fedoraproject.org/wiki/Fedora_40_Mass_Rebuild

* Fri Aug 25 2023 Charalampos Stratakis <cstratak@redhat.com> - 2.4.0-1
- Update to 2.4.0
- Fixes: rhbz#2121551

* Fri Jul 21 2023 Fedora Release Engineering <releng@fedoraproject.org> - 2.0.3-4
- Rebuilt for https://fedoraproject.org/wiki/Fedora_39_Mass_Rebuild

* Wed Jun 14 2023 Python Maint <python-maint@redhat.com> - 2.0.3-3
- Rebuilt for Python 3.12

* Fri Jan 20 2023 Fedora Release Engineering <releng@fedoraproject.org> - 2.0.3-2
- Rebuilt for https://fedoraproject.org/wiki/Fedora_38_Mass_Rebuild

* Thu Aug 11 2022 Tomáš Hrnčiar <thrnciar@redhat.com> - 2.0.3-1
- Update to 2.0.3
- Fixes: rhbz#2022465

* Fri Jul 22 2022 Fedora Release Engineering <releng@fedoraproject.org> - 1.7.0-4
- Rebuilt for https://fedoraproject.org/wiki/Fedora_37_Mass_Rebuild

* Mon Jun 13 2022 Python Maint <python-maint@redhat.com> - 1.7.0-3
- Rebuilt for Python 3.11

* Fri Jan 21 2022 Fedora Release Engineering <releng@fedoraproject.org> - 1.7.0-2
- Rebuilt for https://fedoraproject.org/wiki/Fedora_36_Mass_Rebuild

* Wed Aug 18 2021 Charalampos Stratakis <cstratak@redhat.com> - 1.7.0-1
- Update to 1.7.0 (#1989154)

* Mon Jul 26 2021 Tomas Hrnciar <thrnciar@redhat.com> - 1.6.4-1
- Update to 1.6.4

* Fri Jul 23 2021 Fedora Release Engineering <releng@fedoraproject.org> - 1.4.2-3
- Rebuilt for https://fedoraproject.org/wiki/Fedora_35_Mass_Rebuild

* Wed Jun 02 2021 Python Maint <python-maint@redhat.com> - 1.4.2-2
- Rebuilt for Python 3.10

* Thu Jan 28 11:17:16 CET 2021 Tomas Hrnciar <thrnciar@redhat.com> - 1.4.2-1
- Update to 1.4.2

* Wed Jan 27 2021 Fedora Release Engineering <releng@fedoraproject.org> - 1.4.1-2
- Rebuilt for https://fedoraproject.org/wiki/Fedora_34_Mass_Rebuild

* Thu Jan 14 08:42:36 CET 2021 Tomas Hrnciar <thrnciar@redhat.com> - 1.4.1-1
- Update to 1.4.1

* Mon Sep 21 2020 Lumír Balhar <lbalhar@redhat.com> - 1.0.18-5
- Fix FTBFS by build-requiring python3-devel

* Wed Jul 29 2020 Fedora Release Engineering <releng@fedoraproject.org> - 1.0.18-4
- Rebuilt for https://fedoraproject.org/wiki/Fedora_33_Mass_Rebuild

* Sat May 23 2020 Miro Hrončok <mhroncok@redhat.com> - 1.0.18-3
- Rebuilt for Python 3.9

* Thu Jan 30 2020 Fedora Release Engineering <releng@fedoraproject.org> - 1.0.18-2
- Rebuilt for https://fedoraproject.org/wiki/Fedora_32_Mass_Rebuild

* Tue Dec 31 2019 Miro Hrončok <mhroncok@redhat.com> - 1.0.18-1
- Update to 1.0.18

* Tue Dec 17 2019 Miro Hrončok <mhroncok@redhat.com> - 1.0.16-1
- Initial package
