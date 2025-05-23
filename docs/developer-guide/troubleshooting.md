# Edge Microvisor Toolkit Troubleshooting

Each build artifact is accompanied with a SHA256 checksum file
that can be used to ensure the integrity of the original image file.

This section provides additional tips and suggestions covering common questions
and issues users may run into.


````{dropdown} How do I verify the integrity of a microvisor image?
Each build artifact is accompanied with a SHA256 checksum file
that can be used to ensure the integrity of the original image file.
Linux
```bash
sha256sum -c full-3.0.20250320.0243.iso.sha256sum
20250320.0243/full-3.0.20250320.0243.iso: OK
# Manually
sha256sum 20250320.0243/full-3.0.20250320.0243.iso
85b137474ec0b9bd79f8573af56f80888cef2f34e6c0649da79dfb58aa28a3bb  20250320.0243/full-3.0.20250320.0243.iso
```

> **Note:** If using `sha256sum` with the -c option, ensure that the image is located in the expected directory. In this case /20250320.0243.

Windows
```winbatch
> certutil -hashfile full-3.0.20250320.0243.iso SHA256
SHA256 hash of full-3.0.20250320.0243.iso:
85b137474ec0b9bd79f8573af56f80888cef2f34e6c0649da79dfb58aa28a3bb
CertUtil: -hashfile command completed successfully.
```
````

````{dropdown} Edge Microvisor Toolkit does not boot. I'm trying to run it as a VM.
There are several reasons why the image may not successfully boot.
Follow these steps to ensure booting is successful.
- Hypervisors may not support custom platform keys for secure boot. If secure
boot is enabled the VM will not boot as it cannot verify the signature. Verify
that secure boot is disabled in your hypervisor settings.
- Inspect that the boot order is configured properly. If you have run the ISO
installer, verify that the VHD (Virtual Hard Drive) has the highest priority so the
VM does not attempt to boot from USB or PXE (Network Boot)
- Make sure EFI is enabled and supported by the hypervisor under settings.
````


````{dropdown} How many CPUs and how much memory should I allocate?
While Edge Microvisor Toolkit will work with constrained devices or VMs,
it is recommended that you have at least 4 CPUs and 4096MB of memory. For quick
testing purposes, you can boot and evaluate Edge Microvisor Toolkit with as little as
a single CPU core with 1024MB of memory.
````


````{dropdown} How much CPU and memory does Edge Microvisor Toolkit consume when idle?
With a *single* CPU core you should see idle consumption
to be around 300MB of memory and on average 3-7% CPU utilization.
````


````{dropdown} How much disk space do I need to allocate for Edge Microvisor Toolkit?
It really depends on your usage scenarios, but Edge Microvisor Toolkit
requires about 750MB of disk space for the root filesystem. The VHD and RAW
images are preallocated with 2GB across the partitions.
````


````{dropdown} How do I install additional rpm packages?
The immutable images (RAW, VHD) have a read-only filesystem and do
not allow you to install any additional packages to the root filesystem.
- You can install both the ISO image and additional rpm packages from the Open
Edge repository that is pre-configured in the image. There are over 3000
packages available.
- If you need to get additional packages enabled in the immutable image you need
to add the rpm(s) to an image file. You can use an existing image or create a new one.
- If a SPEC file does not exist for the rpm you want to add to your image,
you need to create it as well.
````


````{dropdown} I have an rpm that is not available in Open Edge repository, how do I install it?
If you are using the ISO build you can configure additional
repositories by creating a repository configuration file.
```bash
[custom-repo]
name=My Custom Repository
baseurl=https://example.com/path/to/repo
enabled=1
gpgcheck=1
gpgkey=https://example.com/path/to/RPM-GPG-KEY-myrepo
```
- **enabled**: Set to `1` to enable the repo
- **gpgcheck**: Set to `1` to enable GPG signature verification
Finally refresh the tdnf cache
```bash
tdnf clean all
tdnf makecache
```
````


````{dropdown} Will my home directory be saved if I perform an update of Edge Microvisor Toolkit?
Yes, the entire `/home` directory is configured as a persistent
bind mount and will be kept across updates. This is true for other key
directories as well.
````


````{dropdown} Is virtualization supported in Edge Microvisor Toolkit?
Yes, Edge Microvisor Toolkit offers native support for virtualization through
KVM/qemu. You can install it as a bare metal/host operating
system and run guest operating systems on top.
````


````{dropdown} Is SRIOV supported in Edge Microvisor Toolkit?
Yes, SRIOV is supported at the kernel level in Edge Microvisor Toolkit.
````


````{dropdown} Does Edge Microvisor Toolkit support NVIDIA GPUs?
NVIDIA GPUs are not currently supported.
````


````{dropdown} Are the microvisor images signed?
Yes, the production images are signed by Intel, and because the UKI is
signed, BIOS needs to be configured with those keys
````


````{dropdown} How do I add a Kernel module (.ko) file?
It depends on which microvisor image you are using.

* Mutable ISO: You can add or update modules at runtime. Standard commands like
`insmod` or `modprobe` enable you to load a downloaded or newly built .ko file.
The file system is writable, plus the `dm‑verity` feature is not enabled,
so you can modify or add kernel modules as needed.
* Immutable Image: The OS image is read-only; you cannot download or add new .ko
files after deployment. If a module is needed, it must be included in the system
image at build time. Assuming the module is a part of the verified image, you can
load it using the usual methods (`insmod` or `modprobe`). `dm‑verity` ensures the
integrity of the image, so any module loaded must match the signed, verified
version in the image.
````

````{dropdown} How do I change the kernel command line (e.g. Huge Pages)?
It depends on which image of the Edge Microvisor you are using.

* Mutable ISO: You can modify the kernel command line by updating the bootloader
configuration (`/etc/default/grub`) and then regenerate the configuration by
running `grub-mkconfig`.
* Immutable Image: You cannot change the kernel line currently without rebuilding
the image as the command line is included in the UKI and signed. The kernel
command line can be modified in the image file (e.g. `edge-image.json`). The
`KernelCommandLine` JSON attribute can be updated to include desired kernel command
line parameter(s).
````

````{dropdown} What do the many JSON files in imageconfigs do? Which needs to be modified for the ISO or the immutable Edge Microvisor Toolkit?
Most of this is captured in the [Building Howto](./get-started/building-howto.md).
The `imageconfigs` includes a set of different image files that defines different
image types the Buildkit can produce. For different validated images, Edge Microvisor Toolkit uses the `edge-image.json` and
`edge-image-rt.json`, and image types the Buildkit can produce
````

````{dropdown} What changes are needed on the BIOS to support Edge Microvisor Toolkit?
To install from a USB device, BIOS needs to be updated to include the
USB boot option and you need to make sure USB boot has highest precedence in the
boot order list. You would also need to configure BIOS with the Platform Keys to enable
[secure boot](./get-started/sb-howto.md) for Edge Microvisor Toolkit.
````