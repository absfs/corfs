# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive unit tests for all File and FileSystem methods
- Benchmark tests for performance validation
- Code coverage improvements (80%+ coverage)
- Additional badges in README (Go Reference, Go Report Card, Build Status)
- CHANGELOG.md for version tracking
- CONTRIBUTING.md for contributor guidelines
- CODE_OF_CONDUCT.md for community standards

### Fixed
- Code formatting issues in test files
- Missing `os` import in README example

## [0.1.0] - 2024-11-08

### Added
- Initial implementation of Cache-on-Read FileSystem
- Two-tier caching system with primary and cache filesystems
- Full implementation of `absfs.Filer` interface
- Basic unit tests
- CI/CD with GitHub Actions
- Documentation and README
- MIT License

[Unreleased]: https://github.com/absfs/corfs/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/absfs/corfs/releases/tag/v0.1.0
