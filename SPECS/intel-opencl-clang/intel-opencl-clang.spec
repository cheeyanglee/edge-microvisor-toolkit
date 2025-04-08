%global llvm_compat 14
%global commit 470cf0018e1ef6fc92eda1356f5f31f7da452abc
%global shortcommit %(c=%{commit}; echo ${c:0:8})

Name:           intel-opencl-clang
Version:        140
Release:        2%{?dist}
Summary:        Library to compile OpenCL C kernels to SPIR-V modules
License:        Apache-2.0 WITH LLVM-exception OR NCSA
Vendor:         Intel Corporation
Distribution:   Edge Microvisor Toolkit
URL:            https://github.com/intel/opencl-clang
Source0:        %{url}/archive/%{commit}/%{name}-%{shortcommit}.tar.gz

BuildRequires:  cmake
BuildRequires:  clang%{?llvm_compat}
BuildRequires:  gcc
BuildRequires:  gcc-c++
BuildRequires:  make
BuildRequires:  llvm%{?llvm_compat}-devel
BuildRequires:  clang%{?llvm_compat}-devel
BuildRequires:  spirv-llvm-translator-devel
BuildRequires:  spirv-llvm-translator
BuildRequires:  zlib-devel

%description
opencl-clang is a thin wrapper library around clang. The library has OpenCL-oriented API and
is capable to compile OpenCL C kernels to SPIR-V modules.

%package devel
Summary: Development files for %{name}
Requires: %{name}%{?_isa} = %{version}-%{release}

%description devel
This package contains libraries and header files for
developing against %{name}

%prep
%autosetup -n opencl-clang-%{commit} -p1

%build
%cmake \
    -DLLVM_DIR=%{_libdir}/cmake/llvm/ -DCMAKE_BUILD_TYPE=Release
%cmake_build

%install
%cmake_install

%files
%license LICENSE
%{_libdir}/libopencl-clang.so.*

%files devel
%{_libdir}/libopencl-clang.so
%{_includedir}/cclang/common_clang.h
%{_includedir}/cclang/opencl-c.h
%{_includedir}/cclang/opencl-c-base.h
%{_includedir}/cclang/module.modulemap

%changelog
* Tue Dec 24 2024 Naveen Saini <naveen.kumar.saini@intel.com> - 140-2
- Updated initial changelog entry having fedora version and license info.

* Fri Sep 27 2024 Junxiao Chang <junxiao.chang@intel.com> - 140-1
- Initial Edge Microvisor Toolkit import from Fedora 36 (license: MIT). License verified.
