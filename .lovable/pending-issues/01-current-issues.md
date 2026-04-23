# Pending Issues

## 01 — Unit Test Coverage Gaps
- **Status**: Open since v2.49.0
- **Description**: Missing unit tests for `task`, `env`, and `install` command families
- **Impact**: Low — commands work but lack automated regression coverage
- **Blocked By**: Nothing — can be done anytime
- **Files Affected**: `cmd/task*.go`, `cmd/env*.go`, `cmd/install*.go`

## 02 — Install --check Missing "Not Found" Message
- **Status**: Open since v2.49.0
- **Description**: `gitmap install --check <tool>` doesn't print a distinct message when a tool is not installed; constant was added but wiring is incomplete
- **Impact**: Low — tool still works, just poor UX for missing tools
- **Files Affected**: `cmd/installtools.go`

## 03 — Docs Site Navigation Missing Pages
- **Status**: Open since v2.76.0
- **Description**: `version-history` and `clone` pages exist but are not linked from the sidebar or commands page navigation
- **Impact**: Low — pages exist at `/version-history` and users won't discover them organically
- **Files Affected**: Sidebar component, `src/data/commands.ts`

## 04 — Helptext/env.md Missing --shell Examples
- **Status**: Open since v2.49.0
- **Description**: The `--shell` flag was wired into env commands but the help text file doesn't demonstrate usage
- **Impact**: Low — flag works but users won't know about it from `gitmap help env`
- **Files Affected**: `helptext/env.md`

## 05 — Clone-Next Missing --dry-run Support
- **Status**: Open (feature gap)
- **Description**: The flatten spec (87-clone-next-flatten.md) mentions `--dry-run` for previewing clone-next actions but it's not implemented
- **Impact**: Medium — users can't preview destructive folder removal before it happens
- **Files Affected**: `cmd/clonenext.go`, `cmd/clonenextflags.go`, `constants/constants_clonenext.go`
