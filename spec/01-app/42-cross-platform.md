# Cross-Platform Support

## Overview

gitmap supports Windows, Linux, and macOS with feature parity across
all platforms. Build scripts, deployment, and CI/CD are designed to
work identically regardless of the host OS.

---

## Build Scripts

### `run.ps1` (Windows)

PowerShell script for Windows environments. Primary development script
with full pipeline support: pull → build → deploy → run.

### `run.sh` (Linux / macOS)

Bash script mirroring `run.ps1` with identical steps and output format.
Reads configuration from the same `powershell.json` file using `jq` or
`python3` for JSON parsing.

### Parity

| Feature                        | run.ps1 | run.sh |
|--------------------------------|---------|--------|
| Git pull with branch check     | ✅      | ✅     |
| Dependency resolution          | ✅      | ✅     |
| Source file validation         | ✅      | ✅     |
| Build with ldflags             | ✅      | ✅     |
| Data folder copy               | ✅      | ✅     |
| Smart deploy path resolution   | ✅      | ✅     |
| Rename-first deploy            | ✅      | ✅     |
| Deploy with retry              | ✅      | ✅     |
| Rollback on failure            | ✅      | ✅     |
| PATH sync detection            | ✅      | ✅     |
| Test mode (`-t`)               | ✅      | ✅     |
| Colored output                 | ✅      | ✅     |

---

## Makefile

The `gitmap/Makefile` provides a standard interface wrapping `run.sh`:

| Target       | Description                    | Command               |
|--------------|--------------------------------|-----------------------|
| `build`      | Full pipeline (pull+build+deploy) | `bash ../run.sh`   |
| `run`        | Build and run with `ARGS=`     | `bash ../run.sh -r`   |
| `test`       | Run unit tests with reports    | `bash ../run.sh -t`   |
| `update`     | Pull, build, deploy, sync PATH | `bash ../run.sh --update` |
| `no-pull`    | Build without git pull         | `bash ../run.sh --no-pull` |
| `no-deploy`  | Build without deploying        | `bash ../run.sh --no-deploy` |
| `clean`      | Remove build artifacts         | `rm -rf bin/`         |
| `help`       | Show available targets         | grep + awk            |

---

## Platform Detection

The build script auto-detects the platform:

```bash
BINARY_NAME="gitmap"
if [[ "$(uname -s)" == *MINGW* ]] || [[ "$(uname -s)" == *MSYS* ]]; then
    BINARY_NAME="gitmap.exe"
fi
```

File size display uses `stat -f%z` on macOS and `stat -c%s` on Linux.

---

## Cross-Compilation

The Go toolchain builds static binaries for all platforms using:

```
CGO_ENABLED=0 GOOS=<os> GOARCH=<arch> go build
```

Default targets (6):

| OS      | Arch  |
|---------|-------|
| windows | amd64 |
| windows | arm64 |
| linux   | amd64 |
| linux   | arm64 |
| darwin  | amd64 |
| darwin  | arm64 |

Targets are configurable via `release.targets` in `config.json` or
the `--targets` CLI flag.

---

## CI/CD — GitHub Actions

### Workflow: `.github/workflows/ci.yml`

Triggers on push to `main` and pull requests.

**Jobs:**

1. **test** — Runs `go test ./...` on ubuntu-latest
2. **build** — Cross-compiles all 6 targets, verifying each binary
3. **release** — On tagged pushes (`v*`), creates a GitHub Release
   with compressed assets and checksums

### Workflow: `.github/workflows/release.yml`

Triggers on tags matching `v*`.

**Steps:**

1. Checkout code
2. Set up Go toolchain
3. Run `gitmap release` with `--compress --checksums`
4. Upload release assets to GitHub

---

## Constraints

- All scripts must read from `powershell.json` as the single source of truth
- Bash scripts require `bash 4+` (macOS ships `bash 3` — use `#!/usr/bin/env bash`)
- Binary names: `gitmap` on Unix, `gitmap.exe` on Windows
- No external CLI dependencies beyond `git`, `go`, `jq` or `python3`

## Cross-References (Generic Specifications)

| Topic | Generic Spec | Covers |
|-------|-------------|--------|
| Build scripts | [04-build-scripts.md](../08-generic-update/04-build-scripts.md) | `run.ps1` / `run.sh` pipeline, config loading, platform detection |
| PowerShell patterns | [02-powershell-build-deploy.md](../03-general/02-powershell-build-deploy.md) | Script architecture, step-based execution |
| Deploy strategy | [03-rename-first-deploy.md](../08-generic-update/03-rename-first-deploy.md) | Rename-first on Windows, `mv` on Unix |
| Cross-compilation | [01-cross-compilation.md](../07-generic-release/01-cross-compilation.md) | Multi-platform Go build targets |
