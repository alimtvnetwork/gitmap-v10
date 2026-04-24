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

## 06 — Multi-URL Clone: PowerShell Comma-Splitting Crash (FIXED v3.80.0)
- **Status**: Fixed in v3.80.0
- **Reported**: User ran `gitmap clone url1,url2,url3` in PowerShell on Windows; got `fatal: could not create leading directories of 'D:\...\https:\github.com\alimtvnetwork\email-reader-v3.gitmap-tmp-...': Invalid argument`
- **Root Cause**:
  1. PowerShell on Windows silently splits unquoted comma-separated arguments into multiple `argv` entries when invoking external executables. So `url1,url2,url3` arrived as three separate `os.Args` entries, not one string.
  2. `parseCloneFlags` only inspected the first two positional args: `Arg(0)` became the source URL, `Arg(1)` was treated as the **folder name**.
  3. `executeDirectClone` then called `filepath.Abs("https://github.com/.../email-reader-v3")`, producing the nonsense Windows path `D:\...\https:\github.com\alimtvnetwork\email-reader-v3` (illegal because `:` is reserved after the drive letter).
  4. The replace-strategy code then tried to `os.RemoveAll` and `git clone` into that path, both of which fail with "filename, directory name, or volume label syntax is incorrect" / "could not create leading directories".
  5. Spec `01-app/104-clone-multi.md` and `mem://features/clone-multi` had been **planned for v3.38.0 but never implemented** — the parser still assumed exactly one source.
- **Solution**:
  1. New `flattenURLArgs([]string) []string` (`gitmap/cmd/clonemulti.go`) — splits each positional arg on `,`, trims whitespace, drops empties, dedupes case-insensitively (normalising trailing `.git`), preserving first-seen order. Accepts both `a b c` and `a,b,c` and mixed `a,b c d,e`.
  2. `parseCloneFlags` now returns a `CloneFlags` struct exposing the **full positional slice** (not just `Arg(0)`/`Arg(1)`).
  3. `resolveCloneFolderName` defensively returns `""` when the second positional arg looks like a URL — so even single-URL invocations can't be misinterpreted as `<url> <folder=other-url>`.
  4. `runClone` detects multi-URL form (any positional contains `,`, or 2+ positionals where both Arg(0) and Arg(1) parse as URLs) and dispatches to the new `runCloneMulti` worker which calls a non-fatal `executeDirectCloneOne` per URL, continuing on failure.
  5. Exit codes per spec: `0` all OK, `1` partial failure, `3` all URLs invalid.
- **Files Affected**:
  - `gitmap/cmd/clone.go` — new `runClone` dispatcher + `shouldUseMultiClone` + `runCloneMulti`
  - `gitmap/cmd/clonemulti.go` (new) — `flattenURLArgs`, `classifyURLs`, `executeDirectCloneOne`, `resolveCloneFolder`, `normaliseURLKey`
  - `gitmap/cmd/rootflags.go` — `CloneFlags` struct, `isLikelyURL` guard
  - `gitmap/constants/constants_clone.go` — `MsgCloneInvalidURLFmt`, `MsgCloneSummaryMultiFmt`, `MsgCloneRegisteredInline`, `MsgCloneMultiBegin`, `MsgCloneMultiItem`, `ErrCloneAllInvalid`, `ErrCloneMultiFailedFmt`, `ExitCloneMultiPartialFail`, `ExitCloneMultiAllInvalid`
  - `gitmap/constants/constants.go` — version bumped to `3.80.0`
- **PowerShell Note**: Even after this fix, users should prefer space-separated URLs in PowerShell to avoid relying on PS's implicit comma-splitting (which differs across PS 5.1 / 7.x). Both forms now work either way.
