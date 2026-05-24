<pre align="center">
___  ___        ___  ________  ________  ___       ___  _________   _______     
|\  \|\  \      |\  \|\   ____\|\   __  \|\  \     |\  \|\___   ___\/  ___  \    
\ \  \\\  \     \ \  \ \  \___|\ \  \|\  \ \  \    \ \  \|___ \  \_/__/|_/  /|   
 \ \   __  \  __ \ \  \ \_____  \ \   ____\ \  \    \ \  \   \ \  \|__|//  / /   
  \ \  \ \  \|\  \\_\  \|____|\  \ \  \___|\ \  \____\ \  \   \ \  \   /  /_/__  
   \ \__\ \__\ \________\____\_\  \ \__\    \ \_______\ \__\   \ \__\ |\________\
    \|__|\|__|\|________|\_________\|__|     \|_______|\|__|    \|__|  \|_______|
                        \|_________|
</pre>

<p align="center">
  <strong>A spiritual successor to HJSplit</strong><br>
  <em>Fast, cross-platform file splitting and joining tool</em>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go" alt="Go version">
  <img src="https://img.shields.io/badge/Platform-Windows%20|%20Linux%20|%20macOS-blue" alt="Platform">
  <img src="https://img.shields.io/badge/License-GPLv3-red" alt="License">
  <img src="https://img.shields.io/badge/GUI-Fyne-cyan" alt="GUI">
  <img src="https://img.shields.io/badge/status-stable-brightgreen" alt="Status">
</p>

---

## Overview

**hjsplit2** is a modern, open-source replacement for the classic HJSplit utility. It lets you split large files into smaller chunks and rejoin them later — perfect for transferring big files over email, USB drives, or cloud storage with size limits.

Built with **Go** and the **Fyne** GUI toolkit, featuring a dark navy/synthwave theme inspired by classic terminal aesthetics.

## Features

- ✂️ **Split** — divide any file into custom-sized parts (`.001`, `.002`, ...)
- 🔗 **Join** — merge split files back to the original
- 🔀 **Custom Join** — join multiple files in any order via drag-and-drop list
- 🖱️ **Drag & Drop** — drop files directly onto the window
- 🎨 **Synthwave Theme** — dark navy, hot pink, and teal aesthetics
- 🖥️ **Console Log** — real-time operation log with terminal-style output
- 📦 **Portable** — single executable, no installation required
- 🌍 **Cross-platform** — Windows, Linux, macOS

## Downloads

Pre-built binaries for Windows (32-bit and 64-bit) and Linux are available in the [dist](./dist) directory after building.

| Platform | File |
|----------|------|
| Windows 64-bit | `hjsplit2-v*-windows-amd64.exe` |
| Windows 32-bit | `hjsplit2-v*-windows-386.exe` |
| Linux 64-bit | `hjsplit2-v*-linux-amd64` |

## Building

### Prerequisites

- [Go](https://go.dev/dl/) 1.21 or later
- [MinGW-w64](https://winlibs.com/) (Windows, for CGO/GLFW)
- [UPX](https://upx.github.io/) (optional, for smaller binaries)

### Quick build (native)

```bash
go build -ldflags="-s -w -H windowsgui" -o hjsplit2.exe .
```

### Using build scripts

```powershell
# Windows — build all platforms
.\build.ps1 all

# Build only for current platform
.\build.ps1 native

# Build specific targets
.\build.ps1 win64
.\build.ps1 win32         # requires MINGW32_PATH in build.conf
.\build.ps1 linux          # requires WSL with Go installed
```

```bash
# Linux / WSL
./build.sh all
./build.sh native
./build.sh win64           # requires MinGW cross-compiler
```

### Cross-compilation notes

The build scripts automatically detect available toolchains:

- **win64**: uses system MinGW-w64 (64-bit)
- **win32**: configure `MINGW32_PATH` or `CC_386` in `build.conf`
- **Linux (from Windows)**: builds via WSL if Go is installed in WSL
- **Linux (native)**: builds directly

Configure paths in [`build.conf`](./build.conf) (copy from `build.conf.example`):

```ini
MINGW32_PATH=C:/Apps/mingw32
WSL_DISTRO=Ubuntu
USE_UPX=true
```

## Usage

### Split
1. Go to the **Split** tab
2. Drag a file onto the window (or click to browse)
3. Set the chunk size (KB, MB, or GB)
4. Click **Split**

### Join (auto-detect)
1. Go to the **Join** tab
2. Drag a `.001` file onto the window
3. Click **Join** — automatically finds and merges `.001`, `.002`, ...

### Custom Join (any order)
1. Go to the **Join** tab
2. Click **Custom Join...**
3. Add files with **Add files...**, reorder with **↑ / ↓**
4. Set the output path and click **START JOIN**

Output files from Split follow the `.001`, `.002`, ... naming convention, compatible with the original HJSplit format.

## Tech Stack

- **Language**: [Go](https://go.dev/)
- **GUI**: [Fyne](https://fyne.io/) v2
- **Theme**: Custom Synthwave (Python‑inspired palette)
- **Windowing**: GLFW via Fyne
- **Original concept**: [HJSplit](http://www.hjsplit.org/)

## License

This project is licensed under the **GNU General Public License v3.0** — see the [LICENSE](./LICENSE) file for details.

## Attribution

<a target="_blank" href="https://icons8.com/icon/Msb7gOpzHHvG/split-files">Split</a> icon by <a target="_blank" href="https://icons8.com">Icons8</a>
