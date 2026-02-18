# heic2jpg

> Convert HEIC photos to JPEG with a single command. No installation required.

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A fast, simple command-line tool to convert HEIC images to JPEG format while preserving EXIF metadata.

## Quick Start

```bash
# Convert all HEIC files in current directory
heic2jpg

# Convert a specific directory
heic2jpg photos/

# Convert recursively with custom output
heic2jpg -r --output ./converted photos/
```

## Features

- 🚀 **Fast conversion** - Efficiently converts HEIC images to high-quality JPEG
- 📸 **EXIF preservation** - Maintains photo metadata (date, location, camera settings)
- 🎯 **Flexible naming** - Supports custom output patterns with placeholders
- 📁 **Batch processing** - Convert entire directories at once
- 🔍 **Dry-run mode** - Preview conversions before executing
- 💻 **Cross-platform** - Works on macOS, Linux, and Windows

## Installation

### Option 1: Download Binary

Go to the [Releases page](https://github.com/yishaym/heic2jpg/releases/latest) and download the binary for your platform:

| Platform | File |
|----------|------|
| macOS (Apple Silicon) | `heic2jpg-darwin-arm64` |
| macOS (Intel) | `heic2jpg-darwin-amd64` |
| Linux | `heic2jpg-linux-amd64` |
| Windows | `heic2jpg-windows-amd64.exe` |

**Install on macOS/Linux:**
```bash
chmod +x heic2jpg-*
sudo mv heic2jpg-* /usr/local/bin/heic2jpg
```

**Install on Windows:**
1. Download `heic2jpg-windows-amd64.exe` from the [Releases page](https://github.com/yishaym/heic2jpg/releases/latest)
2. Rename it to `heic2jpg.exe` (optional, for easier typing)
3. Move it to a folder like `C:\Tools\`
4. Add that folder to your PATH:
   - Press `Win + R`, type `sysdm.cpl`, press Enter
   - Go to **Advanced** → **Environment Variables**
   - Under "User variables", select **Path** → **Edit** → **New**
   - Add `C:\Tools` (or wherever you put the file)
   - Click **OK** on all windows
5. Open a new Command Prompt and type `heic2jpg` to verify it works

### Option 2: Build from Source

Requirements: Go 1.21+, C compiler

```bash
git clone https://github.com/yishaym/heic2jpg.git
cd heic2jpg
make build
```

The binary will be in the `dist/` directory.

**Install the binary:**
```bash
# macOS/Linux
chmod +x dist/heic2jpg-*
sudo mv dist/heic2jpg-* /usr/local/bin/heic2jpg

# Or add to your PATH
export PATH=$PATH:$(pwd)/dist
```

## Usage

### Basic Usage

Convert all HEIC files in the current directory:
```bash
heic2jpg
```

Convert all HEIC files in a specific directory:
```bash
heic2jpg /path/to/photos
```

Convert a single file:
```bash
heic2jpg photo.heic
```

### Advanced Options

**Custom output directory:**
```bash
heic2jpg --output ./converted photos/
```

**Custom naming pattern:**
```bash
# Add suffix to original names
heic2jpg --pattern "{name}_converted"

# Sequential numbering
heic2jpg --pattern "IMG_{index}"

# Include date in filename
heic2jpg --pattern "{date}_{name}"
```

**Recursive directory scanning:**
```bash
heic2jpg --recursive photos/
# or use short form
heic2jpg -r photos/
```

**Dry-run mode (preview without converting):**
```bash
heic2jpg --dry-run photos/
```

**Skip confirmation prompts:**
```bash
# Useful for automation or when processing many files
heic2jpg --yes photos/
# or use short form
heic2jpg -y photos/
```

**Overwrite existing files:**
```bash
heic2jpg --force photos/
# or use short form
heic2jpg -f photos/
```

**Limit file size:**
```bash
# Only process files up to 500MB (default)
heic2jpg --max-size 500MB photos/

# Process files up to 1GB
heic2jpg --max-size 1GB photos/

# Disable size limit (use with caution)
heic2jpg --max-size 0 photos/
```

**Strict EXIF mode:**
```bash
# Fail conversion if EXIF metadata is missing
heic2jpg --strict photos/
```

**Show version:**
```bash
heic2jpg --version
```

### Naming Patterns

The `--pattern` flag supports the following placeholders:

- `{name}` - Original filename without extension
- `{date}` - File modification date (YYYY-MM-DD format)
- `{index}` - Sequential number (001, 002, 003...)

**Examples:**
```bash
# Original: IMG_1234.heic → IMG_1234.jpg
heic2jpg --pattern "{name}"

# Original: IMG_1234.heic → IMG_1234_web.jpg
heic2jpg --pattern "{name}_web"

# Original: IMG_1234.heic → 2024-01-15_IMG_1234.jpg
heic2jpg --pattern "{date}_{name}"

# Original: IMG_1234.heic → photo_001.jpg
heic2jpg --pattern "photo_{index}"
```

## Examples

**Convert vacation photos:**
```bash
heic2jpg --output ~/Pictures/Converted ~/Pictures/Vacation2024/
```

**Rename with dates for organization:**
```bash
heic2jpg --pattern "{date}_{name}" --output ./organized ./photos/
```

**Preview before converting:**
```bash
heic2jpg --dry-run --pattern "IMG_{index}" ./photos/
```

## Technical Details

- **Image Quality:** JPEG encoding at 95% quality
- **EXIF Support:** Preserves all EXIF metadata from HEIC files
- **Dependencies:** Uses [goheif](https://github.com/jdeng/goheif) for HEIC decoding

## Building

### Build for Current Platform
```bash
make build
```

### Build for All Platforms
```bash
make build-all
```

This attempts to create binaries in the `dist/` directory for:
- macOS (Intel and Apple Silicon)
- Linux (x86_64)
- Windows (x86_64)

**Important Note on Cross-Compilation:**

This project uses CGO (via the goheif library which depends on libde265). Cross-compilation with CGO is complex and requires platform-specific C toolchains and libraries.

**For production releases**, we recommend building natively on each target platform:

- **On macOS:** Run `make build` to create the macOS binary
- **On Linux:** Run `make build` to create the Linux binary
- **On Windows:** Run `make build` to create the Windows binary

The `make build-all` target will attempt cross-compilation but will gracefully skip platforms that fail, ensuring at least the native platform builds successfully.

### Clean Build Artifacts
```bash
make clean
```

## Troubleshooting

### "No HEIC files found"
- Ensure the directory contains `.heic` or `.HEIC` files
- Check file permissions - the tool needs read access to the files

### "Failed to decode HEIC"
- The HEIC file may be corrupted
- Try opening the file in another application to verify it's valid
- Some HEIC variants may not be supported by the underlying library

### "Permission denied" when writing output
- Check that you have write permissions to the output directory
- If using `--output`, ensure the directory exists or the tool can create it

### Build errors with CGO
- Ensure you have a C compiler installed (gcc on Linux, Xcode Command Line Tools on macOS)
- On Linux, you may need to install development libraries: `sudo apt-get install libde265-dev`
- On macOS, install Xcode Command Line Tools: `xcode-select --install`

### Cross-compilation issues
- Cross-compiling with CGO is complex due to C dependencies
- For best results, build natively on each target platform
- See the "Building" section for platform-specific build instructions

## FAQ

### Does this work with HEIF files?
Yes, HEIC is a variant of HEIF. The tool should work with most HEIF images.

### Is the image quality preserved?
Images are encoded at 95% JPEG quality, which provides excellent quality while reducing file size. The conversion is lossy, as JPEG is a lossy format.

### What happens to the original files?
Original HEIC files are never modified or deleted. The tool only creates new JPEG files.

### Can I convert back from JPEG to HEIC?
No, this tool only converts HEIC → JPEG. Converting JPEG → HEIC would require a different tool.

### Does it preserve photo metadata?
Yes, all EXIF metadata (including date, location, camera settings) is preserved in the output JPEG files.

### Can I use this in a script or automation?
Yes! The tool is designed for command-line use and works great in scripts. Use `--dry-run` to test your commands first.

### What platforms are supported?
macOS (Intel and Apple Silicon), Linux (x86_64), and Windows (x86_64). Other platforms may work if you build from source.

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.
