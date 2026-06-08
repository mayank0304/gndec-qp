<div align="center">

# gndec-qp

### 🚀 Lightning-fast past exam question paper downloads for GNDEC students — terminal + TUI.

A compiled, zero-dependency Go CLI backed by a statically embedded multi-year index map for sub-millisecond lookups. Eliminates manual cloud directory navigation entirely.

</div>

---

## ✨ Features

- **🖥️ Interactive TUI** — Fuzzy search, browse, batch download with progress bars
- **⚡ Instant Lookups** — Static compile-time hash maps
- **📦 Batch Downloads** — Retrieves all available papers for a subject at once
- **📂 Smart Directory Structure** — Auto-groups downloads inside your system's `Downloads/Question Papers/` folder
- **🖥️ Cross-Platform** — Native execution across Windows, macOS, and Linux
- **📄 Auto-Open Flag** — Launches PDFs instantly in your default system viewer
- **💾 Recently Used** — Tracks your last 10 subjects for quick access
- **📶 Offline Cache** — Shows which papers are already downloaded

---

## 📦 Installation

### Direct Pull & Install (any platform)

```bash
# Clone, build, and install in one go
git clone https://github.com/IshpreetSingh8264/gndec-qp.git
cd gndec-qp
make install
```

This clones the repo, compiles the binary, and installs it to `/usr/local/bin/qp`.

### Unix (Linux / macOS) — One-liner via curl

```bash
curl -fsSL https://raw.githubusercontent.com/IshpreetSingh8264/gndec-qp/main/install.sh | bash
```

### Windows — PowerShell

```powershell
powershell -c "iwr -Uri 'https://raw.githubusercontent.com/IshpreetSingh8264/gndec-qp/main/install.ps1' -OutFile install.ps1; .\install.ps1"
```

### Via Go Compiler

Requires the Go runtime environment installed on your machine.

```bash
go install github.com/IshpreetSingh8264/gndec-qp@latest
```

### Standalone Binaries

Go to the [Releases](../../releases) page and download the executable for your platform:

- **macOS (M1/M2/M3):** `qp-mac-arm64`
- **macOS (Intel):** `qp-mac-amd64`
- **Linux (64-bit):** `qp-linux-amd64`
- **Linux (ARM64):** `qp-linux-arm64`
- **Windows (64-bit):** `qp-windows-amd64.exe`

---

## 🚀 Quick Start

### Interactive TUI (recommended)

```bash
# Launch the TUI to search, browse, and download papers
qp

# Launch TUI with a subject pre-filled
qp --code PCIT-114
```

### CLI Mode

```bash
# Download all available papers for a subject
qp --code BCS-403

# Download and automatically launch PDFs in your default viewer
qp --code PCIT-114 --auto
```

### TUI Keybindings

| Key | Action |
|-----|--------|
| `Enter` | Search / Select |
| `↑` / `↓` | Navigate list |
| `Space` | Toggle session selection |
| `a` | Select / Deselect all sessions |
| `d` | Download selected sessions |
| `Esc` | Go back |
| `q` | Return to home / Quit |

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

## 🔧 Build from Source

```bash
git clone https://github.com/IshpreetSingh8264/gndec-qp.git
cd gndec-qp
make build       # build for current platform
make build-all   # cross-compile for all platforms (output in build/)
make install     # build + install to /usr/local/bin
```

---

## 🔄 Workflow Comparison

- **Before:** Open browser, open cloud link, wait for UI loading, click year, click branch, click subject, and download individual files manually one by one.
- **After:** Run `qp` (TUI) or `qp --code BCS-403 --auto` to instantly get and open your papers.

---

## 🤝 Contribution & License

Issues and pull requests are welcome for indexing updates — adding missing subjects, years, or fixing broken asset links.

Licensed under the **MIT License**.
