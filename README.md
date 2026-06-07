<div align="center">

# gndec-qp

### 🚀 Lightning-fast past exam question paper downloads for GNDEC students — directly from the terminal.

A compiled, zero-dependency Go CLI backed by a statically embedded multi-year index map for sub-millisecond lookups. Eliminates manual cloud directory navigation entirely.

</div>

---

## ✨ Features

- **⚡ Instant Lookups** — Static compile-time hash maps
- **📦 Batch Downloads** — Retrieves all available papers for a subject at once
- **📂 Smart Directory Structure** — Auto-groups downloads inside your system's `Downloads/Question Papers/` folder
- **🖥️ Cross-Platform** — Native execution across Windows, macOS, and Linux
- **📄 Auto-Open Flag** — Launches PDFs instantly in your default system viewer

---

## 📦 Installation

### Route A — Standalone Binaries (No Go Required)

Go to the [Releases](../../releases) page of this repository, download the executable for your laptop, and run it directly in your terminal:

- **macOS (M1/M2/M3):** Download `qp-mac-arm64` and run `./qp-mac-arm64 --code BCS-403 --auto`
- **Windows (64-bit):** Download `qp-windows-amd64.exe` and run `.\qp-windows-amd64.exe --code BCS-403 --auto`
- **Linux (64-bit):** Download `qp-linux-amd64` and run `./qp-linux-amd64 --code BCS-403 --auto`

---

### Route B — Via Go Compiler

Requires the Go runtime environment installed on your machine.

```bash
go install github.com/mayank0304/gndec-qp@latest
```

> **Note:** Ensure your system environment `PATH` includes your global `go/bin` directory to invoke the command from anywhere:

```bash
gndec-qp --code BCS-403 --auto
```

---

## 🚀 Quick Start

```bash
# 1. Download all available papers for a subject
gndec-qp --code BCS-403

# 2. Download and automatically launch PDFs in your default viewer
gndec-qp --code PCIT-114 --auto
```

---

## 📂 Download Layout

Downloads are mapped dynamically to your system user's default path:

```text
Downloads/
└── Question Papers/
    └── BCS-403/
        ├── 2022.pdf
        ├── 2023.pdf
        └── 2024.pdf
```

---

## 🔄 Workflow Comparison

- **Before:** Open browser, open cloud link, wait for UI loading, click year, click branch, click subject, and download individual files manually one by one.
- **After:** Run a single command `gndec-qp --code BCS-403 --auto` to instantly get and open your papers.

---

## 🤝 Contribution & License

Issues and pull requests are welcome for indexing updates — adding missing subjects, years, or fixing broken asset links.

Licensed under the **MIT License**.
