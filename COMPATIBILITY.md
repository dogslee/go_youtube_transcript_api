# Compatibility Guide

This document provides detailed information about the compatibility requirements and known compatibility issues for the YouTube Transcript API Go implementation.

## Go Version Requirements

### Minimum Required Version

- **Go 1.19.0** or higher is required

The project uses Go 1.19 as the minimum version specified in `go.mod`. This ensures compatibility with all the language features and standard library functions used in the codebase.

### Recommended Version

- **Go 1.21.0** or higher is recommended for the best experience

While the project works with Go 1.19, using a newer version provides:
- Better performance optimizations
- Improved error handling
- Enhanced security features
- Better tooling support

### Testing Matrix

The project is tested against the following Go versions:

| Go Version | Status | Notes |
|------------|--------|-------|
| 1.19.x     | ✅ Supported | Minimum required version |
| 1.20.x     | ✅ Supported | Fully tested |
| 1.21.x     | ✅ Supported | Recommended version |
| 1.22.x     | ✅ Supported | Latest stable version |

## Dependencies

### External Dependencies

| Package | Version | Go Version Requirement | Notes |
|---------|---------|------------------------|-------|
| `github.com/beevik/etree` | v1.1.0 | Go 1.19+ | XML parsing library. v1.6.0+ requires Go 1.21+, so we use v1.1.0 for compatibility |

### Dependency Compatibility

- **etree v1.1.0**: Compatible with Go 1.19 and later versions
  - Note: Newer versions (v1.6.0+) require Go 1.21+ due to use of `iter` and `slices` packages
  - The project uses v1.1.0 to maintain compatibility with Go 1.19

## Operating System Compatibility

The project is compatible with all major operating systems that Go supports:

### Supported Operating Systems

| OS | Status | Notes |
|----|--------|-------|
| Linux | ✅ Fully Supported | Primary development platform |
| macOS | ✅ Fully Supported | Tested on macOS 10.15+ |
| Windows | ✅ Fully Supported | Tested on Windows 10/11 |
| FreeBSD | ✅ Supported | Should work, but not regularly tested |
| OpenBSD | ✅ Supported | Should work, but not regularly tested |
| NetBSD | ✅ Supported | Should work, but not regularly tested |
| DragonFly BSD | ✅ Supported | Should work, but not regularly tested |
| Solaris | ⚠️ Untested | May work, but not tested |
| AIX | ⚠️ Untested | May work, but not tested |

## Architecture Compatibility

The project supports all architectures that Go supports:

### Supported Architectures

| Architecture | Status | Notes |
|--------------|--------|-------|
| amd64 (x86_64) | ✅ Fully Supported | Primary development architecture |
| arm64 (aarch64) | ✅ Fully Supported | Tested on Apple Silicon and ARM servers |
| arm (32-bit) | ✅ Supported | Should work, but not regularly tested |
| 386 (x86) | ✅ Supported | Should work, but not regularly tested |
| ppc64le | ✅ Supported | Should work, but not regularly tested |
| ppc64 | ✅ Supported | Should work, but not regularly tested |
| mips64le | ✅ Supported | Should work, but not regularly tested |
| mips64 | ✅ Supported | Should work, but not regularly tested |
| mipsle | ✅ Supported | Should work, but not regularly tested |
| mips | ✅ Supported | Should work, but not regularly tested |
| riscv64 | ✅ Supported | Should work, but not regularly tested |
| s390x | ✅ Supported | Should work, but not regularly tested |

## Build Tags and Features

### Build Tags

The project uses the following build tags:

- `integration`: Used to mark integration tests that require network access
  - Usage: `go test -tags=integration`
  - These tests are skipped when running with `-short` flag

### Feature Compatibility

All features are available on all supported platforms:

- ✅ Transcript fetching
- ✅ Transcript listing
- ✅ Transcript translation
- ✅ Multiple output formats (JSON, SRT, WebVTT, Text)
- ✅ Proxy support (generic and Webshare)
- ✅ Command-line tool

## Known Compatibility Issues

### 1. Webshare Proxy Configuration

**Issue**: `NewWebshareProxyConfig` internally calls `NewGenericProxyConfig("", "")` which may fail validation.

**Status**: Known issue, may need refactoring in future versions.

**Workaround**: The Webshare proxy configuration uses its own URL generation method, so this doesn't affect functionality.

### 2. Thread Safety

**Issue**: `YouTubeTranscriptApi` is not thread-safe.

**Status**: By design - documented limitation.

**Workaround**: Create separate API instances for each goroutine.

### 3. Cookie Authentication

**Issue**: Cookie authentication is not currently supported.

**Status**: Feature not implemented.

**Impact**: Cannot retrieve transcripts for age-restricted videos.

### 4. IP Blocking

**Issue**: YouTube may block IPs that make frequent requests.

**Status**: External limitation, not a code compatibility issue.

**Workaround**: Use proxies or rotate IPs.

## Version Compatibility

### API Compatibility

The project maintains API compatibility within major versions:

- **v1.x**: All versions are API-compatible
- Breaking changes will result in a new major version (v2.x)

### Backward Compatibility

- Older versions of the library should continue to work with newer Go versions
- Newer versions of the library maintain compatibility with Go 1.19+

## Migration Guide

### Upgrading from Older Versions

If you're using an older version of this library:

1. **Check Go Version**: Ensure you're using Go 1.19.0 or higher
   ```bash
   go version
   ```

2. **Update Dependencies**: Run `go get -u` to update to the latest version
   ```bash
   go get -u github.com/dogslee/youtube_transcript_api
   ```

3. **Test Your Code**: Run your tests to ensure everything still works
   ```bash
   go test ./...
   ```

### Downgrading Dependencies

If you need to use an older version:

1. **Pin the Version**: Use `go get` with a specific version tag
   ```bash
   go get github.com/dogslee/youtube_transcript_api@v1.0.0
   ```

2. **Update go.mod**: The version will be pinned in your `go.mod` file

## Testing Compatibility

### Running Tests

**Unit Tests** (no network required):
```bash
go test -v -short
```

**Integration Tests** (require network):
```bash
go test -v -tags=integration
```

**All Tests**:
```bash
go test -v -tags=integration
```

### Test Compatibility

- Unit tests work on all platforms
- Integration tests require:
  - Network connectivity
  - Access to YouTube (may be blocked in some regions)
  - Stable internet connection

## Platform-Specific Notes

### Linux

- Works on all major distributions
- No special requirements
- Tested on Ubuntu, Debian, CentOS, and Alpine Linux

### macOS

- Works on macOS 10.15 (Catalina) and later
- No special requirements
- Tested on Intel and Apple Silicon (M1/M2/M3)

### Windows

- Works on Windows 10 and Windows 11
- No special requirements
- Tested with both MSYS2/MinGW and native Windows builds

## Getting Help

If you encounter compatibility issues:

1. **Check Go Version**: Ensure you're using Go 1.19.0 or higher
2. **Update Dependencies**: Run `go mod tidy` to ensure dependencies are up to date
3. **Check Issues**: Search existing GitHub issues for similar problems
4. **Create an Issue**: If the problem persists, create a new issue with:
   - Go version (`go version`)
   - Operating system and architecture
   - Error messages or logs
   - Steps to reproduce

## Future Compatibility Plans

### Planned Support

- **Go 1.23+**: Will be supported when released
- **New Architectures**: Will be supported as Go adds support

### Potential Breaking Changes

- Future major versions may require newer Go versions
- Breaking changes will be documented in release notes
- Deprecation warnings will be provided before removal

## References

- [Go Release Notes](https://golang.org/doc/devel/release.html)
- [Go Compatibility Promise](https://golang.org/doc/go1compat)
- [Project GitHub Repository](https://github.com/dogslee/youtube_transcript_api)
- [Original Python Implementation](https://github.com/jdepoix/youtube-transcript-api)

