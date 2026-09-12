# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project
follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [1.0.1] - 2026-09-12

### Changed

- Variadic log arguments are declared as `...any` instead of
  `...interface{}`. The types are identical, so callers are unaffected.
- The package-level `Init` doc comment no longer describes the default
  logger as console-only; it writes to the global log file like every
  other logger.

## [1.0.0] - 2026-09-12

First stable release. From this version on, the public API of the `orchid`
package is covered by semantic versioning: breaking changes will only ship in
a new major version.

### Requirements

- Go 1.24 or newer. Earlier releases declared Go 1.16.

### Breaking changes

- **JSON log lines use lowercase keys**: `severity`, `text`, `module`, `time`.
  Earlier versions wrote `Severity`, `Text`, `Module`, `Time`. Existing files
  appended to by 1.0.0 will contain both spellings.
- **`Configuration.SetDefaultFormat` now returns an `error`** and rejects
  values outside `FormatTXT` and `FormatJSON`, leaving the current format
  unchanged.
- **Colors are no longer always on.** By default they are enabled only when
  stderr is a terminal and the `NO_COLOR` environment variable is unset.
  `Configuration.SetEnableColors` still forces them either way.
- **Console line format**: the stray reset code and space that preceded the
  colored module/severity block have been removed. Lines now start with the
  color code. Output without colors is unchanged.

### Fixed

- A failed `SetLogFile` or `SetDefaultFile` (for example, a path in a
  missing directory) used to leave the path recorded with no open handle,
  so every subsequent log call reported an error and the previous log file
  was lost. The new file is now opened first, and on failure the previous
  configuration stays active.
- File writes from `Logger` instances could race with a concurrent
  `SetLogFile`, `Close`, or `Reset` and write to an already-closed handle,
  silently dropping the line. The configuration lock is now held for the
  whole write.
- After a file write failure, every log call printed an `ORCHID FILE ERROR`
  line to stderr. The first failure of an episode is now reported once, and
  reporting resumes only after a successful write or a reconfiguration.

### Changed

- Path validation now lives in `Configuration.SetDefaultFile` and format
  validation in `Configuration.SetDefaultFormat`, so the public setters
  enforce the same rules as `SetLogFile`, which is now a thin wrapper.
- The 255-byte limit is applied per path component, with a 4096-byte limit
  on the whole path. Deep paths with short components are accepted.
- The redundant package-level mutex around the default logger was removed;
  the logger's own mutex and the configuration lock cover the same cases.
- `examples/basic.go` checks every error, closes the log file on exit, and
  documents that all loggers share one global file.

### Added

- `CHANGELOG.md`, a root `Makefile` (`make check` mirrors CI), and a GitHub
  Actions workflow running vet, race-detector tests, gofmt, staticcheck, and
  the example build on Go 1.24 and stable.
- Test coverage for text and JSON file formats, lowercase JSON keys, color
  and no-color console output, `NO_COLOR` handling, path validation
  boundaries, failed-open recovery, once-per-episode error reporting, and
  instance writes during rapid file swaps. Tests write only to temporary
  directories.

## [0.4.1] - 2025-09-14

- Adds the pkg.go.dev badge to the README.

## [0.4.0] - 2025-09-13

- Global `Configuration` singleton shared by all loggers.
- Thread-safety fixes and a concurrency test suite.
- File handle leak fix and proper file cleanup.
- Fixes the "file name too long" validation error.

## [0.3.0] and earlier

- JSON file output, example program, per-module `Logger` instances,
  and the initial colorized console logger.

[Unreleased]: https://github.com/epiphyte/orchid/compare/v1.0.1...HEAD
[1.0.1]: https://github.com/epiphyte/orchid/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/epiphyte/orchid/compare/v0.4.1...v1.0.0
[0.4.1]: https://github.com/epiphyte/orchid/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/epiphyte/orchid/compare/v0.3.0...v0.4.0
