# Edge Microvisor Toolkit Architecture Diagrams

This directory contains detailed architectural diagrams and flow documentation for the Edge Microvisor Toolkit build system.

## Available Diagrams

### [Build Packages Flow](./build-packages-flow.md)
**Command:** `sudo make build-packages REBUILD_TOOLS=y`

A comprehensive architectural diagram showing the complete build process from repository clone to final RPM generation. This diagram includes:

- **10 Major Phases:**
  1. Prerequisites & Initialization
  2. Makefile Processing
  3. Go Tools Build (27 tools)
  4. Toolchain Stage (download or build)
  5. Chroot Worker Creation
  6. SRPM Packing
  7. Dependency Analysis (graph-based)
  8. Package Build Scheduling
  9. Parallel RPM Building
  10. Output Generation

- **Key Features:**
  - 100+ nodes showing detailed process steps
  - Decision points with branching logic
  - Error handling paths
  - Parallel processing flows
  - Output artifacts at each stage

- **Includes:**
  - Detailed component descriptions
  - Build time estimates
  - Optimization strategies
  - Troubleshooting guide
  - Architectural decision rationale

**Best for:** Understanding the complete build system architecture, debugging build issues, optimizing build performance.

## How to Use These Diagrams

### Viewing Mermaid Diagrams

The diagrams use Mermaid syntax for rendering. You can view them in:

1. **GitHub:** Diagrams render automatically in markdown files
2. **VS Code:** Install "Markdown Preview Mermaid Support" extension
3. **Online:** Copy to https://mermaid.live/
4. **CLI:** Use `mmdc` (mermaid-cli) to export as PNG/SVG

### Reading the Flow

- **Rectangles** = Process steps
- **Diamonds** = Decision points
- **Rounded rectangles** = Start/End points
- **Parallelograms** = Input/Output
- **Colors:**
  - 🔵 Blue = Standard processes
  - 🟡 Yellow = Decisions
  - 🟢 Green = Success states
  - 🔴 Red = Error states
  - 🟣 Purple = Output artifacts

### Following a Build

To trace a specific build scenario:

1. Start at the top "Git Clone Repository" node
2. Follow the path based on your build parameters
3. Reference the "Detailed Component Descriptions" section for more info
4. Check "Common Issues" if you encounter errors

## Quick Reference: Common Build Paths

### Path 1: Standard Full Build
```
Clone → Install Prereqs → Build Go Tools → Download Toolchain → 
Create Chroot → Pack All SRPMs → Analyze Dependencies → 
Build All Packages (parallel) → Generate Outputs
```
**Time:** ~1-2 hours  
**Disk:** ~50GB

### Path 2: Quick Targeted Build
```
Clone → Existing Tools → Existing Toolchain → Existing Chroot → 
Pack Selected SRPMs → Analyze Dependencies → Build Selected Packages → 
Generate Outputs
```
**Command:** `sudo make build-packages SRPM_PACK_LIST="kernel"`  
**Time:** ~10-30 minutes  
**Disk:** ~10GB

### Path 3: Complete From-Scratch Build
```
Clone → Install Prereqs → Build Go Tools → BUILD TOOLCHAIN (2-4 hours) → 
Create Chroot → Pack All SRPMs → Analyze Dependencies → 
Build All Packages → Generate Outputs
```
**Command:** `sudo make build-packages REBUILD_TOOLS=y REBUILD_TOOLCHAIN=y`  
**Time:** ~4-8 hours  
**Disk:** ~80GB

## Build System Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Edge Microvisor Toolkit                   │
│                         Build System                         │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
   ┌────▼────┐          ┌────▼────┐          ┌────▼────┐
   │ Stage 1 │          │ Stage 2 │          │ Stage 3 │
   │Toolchain│          │Packages │          │ Images  │
   └────┬────┘          └────┬────┘          └────┬────┘
        │                     │                     │
   ┌────▼─────────────┐  ┌───▼──────────────┐  ┌──▼───────────┐
   │ • Bootstrap GCC  │  │ • Dependency     │  │ • ISO        │
   │ • Core Libraries │  │   Analysis       │  │ • VHD/VHDX   │
   │ • Build Tools    │  │ • Parallel Build │  │ • RAW        │
   │ • ~200 RPMs      │  │ • ~193 Packages  │  │ • Container  │
   └──────────────────┘  └──────────────────┘  └──────────────┘
```

## Key Build Tools

| Tool | Language | Purpose | Phase |
|------|----------|---------|-------|
| `srpmpacker` | Go | Create source RPMs | SRPM Packing |
| `specreader` | Go | Parse .spec files | Dependency Analysis |
| `grapher` | Go | Build dependency graph | Dependency Analysis |
| `graphpkgfetcher` | Go | Download packages | Dependency Analysis |
| `scheduler` | Go | Orchestrate builds | Package Building |
| `pkgworker` | Go | Execute rpmbuild | Package Building |
| `imagecustomizer` | Go | Customize images | Image Generation |

All tools located in: `toolkit/tools/`

## Build System Files

```
edge-microvisor-toolkit/
├── toolkit/
│   ├── Makefile                 # Main build orchestration
│   ├── scripts/
│   │   ├── pkggen.mk           # Package building logic
│   │   ├── toolchain.mk        # Toolchain management
│   │   ├── chroot.mk           # Build environment
│   │   ├── srpm_pack.mk        # Source RPM creation
│   │   └── imggen.mk           # Image generation
│   ├── tools/                  # Go build tools (27 tools)
│   ├── imageconfigs/           # Image definitions (JSON)
│   └── resources/
│       └── manifests/          # Package lists
├── SPECS/                      # RPM package specs (~193)
├── docs/
│   ├── developer-guide/        # EMT documentation
│   └── diagrams/              # Architecture diagrams (this dir)
└── CLAUDE.md                   # AI assistant guide
```

## Further Reading

- [Get Started Guide](../developer-guide/emt-get-started.md)
- [Architecture Overview](../developer-guide/emt-architecture-overview.md)
- [Build Documentation](../../toolkit/docs/building/building.md)
- [Contribution Guide](../developer-guide/emt-contribution.md)

## Contributing New Diagrams

When adding new architectural diagrams:

1. Use Mermaid syntax for consistency
2. Include detailed descriptions below the diagram
3. Add troubleshooting sections
4. Reference related documentation
5. Update this README with the new diagram

**Diagram naming convention:** `<feature>-<type>.md`
- Example: `image-generation-flow.md`, `toolchain-bootstrap-sequence.md`

---

**Last Updated:** 2026-05-07
