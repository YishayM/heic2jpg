# heic2jpg

A fast, simple command-line tool to convert HEIC images to JPEG format while preserving EXIF metadata.

## Features

- 🚀 **Fast conversion** - Efficiently converts HEIC images to high-quality JPEG
- 📸 **EXIF preservation** - Maintains photo metadata (date, location, camera settings)
- 🎯 **Flexible naming** - Supports custom output patterns with placeholders
- 📁 **Batch processing** - Convert entire directories at once
- 🔍 **Dry-run mode** - Preview conversions before executing
- 💻 **Cross-platform** - Works on macOS, Linux, and Windows

## Installation

### Download Pre-built Binaries

Download the appropriate binary for your platform:

**macOS (Apple Silicon / M1/M2/M3):**
```bash
curl -L -o heic2jpg https://github.com/yourusername/heic2jpg/releases/latest/download/heic2jpg-darwin-arm64
chmod +x heic2jpg
sudo mv heic2jpg /usr/local/bin/
```

**macOS (Intel):**
```bash
curl -L -o heic2jpg https://github.com/yourusername/heic2jpg/releases/latest/download/heic2jpg-darwin-amd64
chmod +x heic2jpg
sudo mv heic2jpg /usr/local/bin/
```

**Linux (x86_64):**
```bash
curl -L -o heic2jpg https://github.com/yourusername/heic2jpg/releases/latest/download/heic2jpg-linux-amd64
chmod +x heic2jpg
sudo mv heic2jpg /usr/local/bin/
```

**Windows (x86_64):**
```powershell
# Download from: https://github.com/yourusername/heic2jpg/releases/latest/download/heic2jpg-windows-amd64.exe
# Add to PATH or run directly
```

### Build from Source

Requirements:
- Go 1.25 or later
- C compiler (for CGO dependencies)

```bash
git clone https://github.com/yourusername/heic2jpg.git
cd heic2jpg
make build
```

The binary will be in the `dist/` directory.

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

**Dry-run mode (preview without converting):**
```bash
heic2jpg --dry-run photos/
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

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
