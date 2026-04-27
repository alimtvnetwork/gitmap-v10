# Clone Pick (sparse-checkout)

Clone only selected paths from a git repo into the current directory
using git's native sparse-checkout. Auto-saves every selection to the
local SQLite database for later replay.

## Alias

cpk

## Usage

    gitmap clone-pick <repo-url> <path1,path2,...> [flags]
    gitmap cpk        <repo-url> <path1,path2,...> [flags]

`<repo-url>` accepts a full HTTPS / SSH URL or `owner/repo` /
`host/owner/repo` shorthand (expanded using `--mode`).

`<paths>` is a comma-separated list of repo-relative folders or files.

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--name <label>` | "" | Save selection under a human label for `--replay` |
| `--branch <name>` | "" | Branch to check out (passed to `git clone --branch`) |
| `--mode <https\|ssh>` | https | URL form for shorthand input |
| `--depth <n>` | 1 | Shallow clone depth (`0` = full history) |
| `--cone` | true | Sparse-checkout cone mode (auto-off for globs/files) |
| `--dest <dir>` | . | Destination directory (created if missing) |
| `--keep-git` | true | Keep `.git` after checkout (`=false` for files-only) |
| `--dry-run` | false | Print plan + git commands without executing |
| `--quiet` | false | Suppress per-step progress on stderr |
| `--force` | false | Allow non-empty `<dest>` |
| `--output <mode>` | (off) | `terminal` prints the standardized branch/from/to/command block on **stdout** before the clone runs. Git progress + the saved-selection line stay on **stderr**. Empty keeps legacy output. |

## Examples

### Example 1: pick a single folder

    gitmap cpk owner/repo docs

**Output:**

    saved selection #4 for github.com/owner/repo ((unnamed))

### Example 2: pick multiple paths

    gitmap cpk owner/repo docs,examples,README.md

### Example 3: dry-run preview

    gitmap clone-pick owner/repo docs --dry-run

**Output:**

    gitmap clone-pick: dry-run
    repo:   https://github.com/owner/repo.git
    dest:   .
    mode:   https   branch: (default)   depth: 1   sparse: cone
    1 path(s) -- pass without --dry-run to actually clone

      + docs

    planned commands:
      $ git clone --filter=blob:none --no-checkout --depth 1 https://github.com/owner/repo.git .
      $ git -C . sparse-checkout set --cone docs
      $ git -C . checkout

### Example 4: files only (no .git)

    gitmap cpk owner/repo docs --keep-git=false --dest snapshot

## Exit codes

    0   success / dry-run ok
    1   runtime failure (git / filesystem / database)
    2   bad CLI usage (missing args, invalid flag value)

## See Also

- [clone](clone.md) -- full repo clone
- [clone-from](clone-from.md) -- clone many repos from a manifest
- [clone-now](clone-now.md) -- re-clone from `gitmap scan` output
