# ezship

**ezship** is a lightweight, high-performance multi-engine container manager designed specifically for Windows users. By leveraging **WSL2** and **Alpine Linux**, it provides a "Docker Desktop" experience without the massive resource overhead.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/wendelmax/ezship)](https://go.dev/)
[![CI](https://github.com/wendelmax/ezship/actions/workflows/ci.yml/badge.svg)](https://github.com/wendelmax/ezship/actions/workflows/ci.yml)
[![Latest Release](https://img.shields.io/github/v/release/wendelmax/ezship?color=green)](https://github.com/wendelmax/ezship/releases)

---

## Key Features

- **Modern TUI Dashboard**: An interactive, visually stunning terminal interface to monitor and manage your infrastructure.
- **Transparent Proxying**: Run `docker ps`, `podman run`, or `kubectl get pods` directly from any Windows terminal.
- **Multi-Engine Support**: Seamlessly switch between **Docker**, **Podman**, **nerdctl (containerd)**, **LXC**, and **k3s**.
- **Real-time Monitoring**: Instant status detection of running engines and versions.
- **Smart Path Translation**: Automatically converts Windows paths (`C:\`) to Linux paths (`/mnt/c/`) for volume mounting.
- **Maintenance Tools**: 
  - `ezship vacuum`: Compresses the WSL disk (`.vhdx`) to reclaim storage space.
  - `ezship prune`: Global cleanup of unused containers, images, and volumes.
  - `ezship reset`: Instantly rebuilds your environment from scratch.

---

## Installation

### One-Line Install
To install **ezship** using our automated PowerShell script, run the following in your terminal:

```powershell
iwr -useb https://raw.githubusercontent.com/wendelmax/ezship/main/scripts/install.ps1 | iex
```

### Manual Download
You can download the pre-compiled binaries for your architecture (AMD64, 386, ARM64) with the embedded icon from the [Releases](https://github.com/wendelmax/ezship/releases) page.

---

## Usage

### Dashboard (TUI)
Simply run `ezship` to open the interactive control panel:
```powershell
ezship
```

### Setup an Engine
Install your favorite container engine inside the Alpine backend:
```powershell
ezship setup docker
# or
ezship setup podman
```

### Transparent Mode
To use `ezship` as a drop-in replacement for `docker`, just rename or copy the binary:
```powershell
cp ezship.exe docker.exe
./docker ps
```

---

## Maintenance

| Command | Description |
| :--- | :--- |
| `ezship prune` | Cleans up unused resources across all engines |
| `ezship vacuum` | Compacts WSL disk file to free up host space |
| `ezship reset` | Completely uninstalls and wipes the ezship environment |

---

## Author

**Jackson Wendel Santos Sá**
- Email: [jacksonwendel@gmail.com](mailto:jacksonwendel@gmail.com)
- GitHub: [@wendelmax](https://github.com/wendelmax)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---
