# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Human-readable error messages for non-HEIC files
- File count confirmation prompt for large batches (skip with `--yes`/`-y`)
- Color-coded error summary at end of batch conversions
- Platform info in `--version` output
- Grouped and documented flags in `--help`
- `[DRY-RUN]` prefix for dry-run output
- Ctrl+C graceful handling with partial summary
- Cancel hint for large batches (10+ files)

### Changed
- Error messages now use lowercase for consistency
- Test assertions updated to match new output format

## [1.0.0] - 2026-02-17

### Added
- Initial release of heic2jpg
- HEIC to JPEG conversion with high-quality output (95% JPEG quality)
- EXIF metadata preservation (date, location, camera settings)
- Directory batch processing - convert entire folders at once
- Flexible naming patterns with placeholders:
  - `{name}` - Original filename without extension
  - `{date}` - File modification date (YYYY-MM-DD format)
  - `{index}` - Sequential numbering (001, 002, 003...)
- Dry-run mode to preview conversions before executing
- Custom output directory support via `--output` flag
- Cross-platform support:
  - macOS (Intel and Apple Silicon)
  - Linux (x86_64)
  - Windows (x86_64)
- Pre-built binaries for all supported platforms
- Comprehensive documentation and examples

### Technical Details
- Built with Go 1.21+
- Uses goheif library for HEIC decoding
- CGO-based implementation for native performance
- Makefile for easy building and cross-compilation

[1.0.0]: https://github.com/yourusername/heic2jpg/releases/tag/v1.0.0

