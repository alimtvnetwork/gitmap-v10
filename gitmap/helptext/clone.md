# gitmap clone

Clone repositories from a structured output file (JSON, CSV, or text),
or clone a single repository directly from a Git URL.

## Alias

c

## Usage

    gitmap clone <source|json|csv|text|url> [folder] [flags]

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| --target-dir \<dir\> | current directory | Base directory for clones |
| --safe-pull | false | Pull existing repos with retry + diagnostics |
| --github-desktop | false | Auto-register with GitHub Desktop (no prompt) |
| --verbose | false | Write detailed debug log |

## Prerequisites

- For file-based clone: run `gitmap scan` first to generate output files
- For URL clone: just provide the HTTPS or SSH URL

## Branch selection strategy

`gitmap clone` decides whether to pass `-b <branch>` to `git clone` based on
each record's `branchSource` (captured during scan):

| branchSource     | Behavior                                                   |
|------------------|------------------------------------------------------------|
| `HEAD`           | Checkout the recorded branch (`-b <branch>`).              |
| `remote-tracking`| Checkout the recorded tracking branch (`-b <branch>`).     |
| `default`        | Checkout the recorded repo default (`-b <branch>`).        |
| `detached`       | Omit `-b`; let the remote's default HEAD decide.           |
| `unknown` / empty| Omit `-b`; let the remote's default HEAD decide.           |

This prevents "Remote branch not found" errors when a scan captured a
detached HEAD or a literal `HEAD` value that cannot be checked out.

## Idempotent clone cache

Repeated `gitmap clone` runs are idempotent. After each successful clone or
pull, gitmap writes a fingerprint (URL, branch, local HEAD SHA, remote HEAD
SHA) to `<target-dir>/.gitmap/clone-cache.json`. On the next run, repos
whose local HEAD still matches the cached SHA — and whose remote tip has
not advanced — are reported as `skipped (cached)` instead of being
re-cloned or re-pulled. If the remote is unreachable (offline), gitmap
trusts the cache as long as the local HEAD still matches.

Delete the cache file to force a full reclone of every entry.

## Examples

### Example 1: Clone from a direct URL (versioned — auto-flattened)

    gitmap clone https://github.com/alimtvnetwork/wp-onboarding-v13.git

**Output:**

    Cloning wp-onboarding-v13 into wp-onboarding...
    Cloned wp-onboarding-v13 successfully.
      + 1 repo(s) added to GitHub Desktop, 0 failed.
      Opening wp-onboarding in VS Code...
      VS Code opened.

### Example 2: Clone URL into a custom folder

    gitmap clone https://github.com/alimtvnetwork/wp-alim.git "my-project"

**Output:**

    Cloning wp-alim into my-project...
    Cloned wp-alim successfully.
      + 1 repo(s) added to GitHub Desktop, 0 failed.
      Opening my-project in VS Code...
      VS Code opened.

### Example 3: Clone from JSON output

    gitmap clone json --target-dir D:\projects

**Output:**

    Cloning from .gitmap/output/gitmap.json...
    [1/12] Cloning my-api... done
    [2/12] Cloning web-app... done
    ...
    Clone complete: 12 succeeded, 0 failed

### Example 4: Clone with safe-pull for existing repos

    gitmap c csv --safe-pull

**Output:**

    [1/8] my-api exists -> pulling... Already up to date.
    [2/8] web-app exists -> pulling... Updated (3 new commits)
    [3/8] Cloning billing-svc... done
    ...
    Clone complete: 8 succeeded, 0 failed

### Example 5: Clone from text file with verbose logging

    gitmap clone text --verbose

**Output:**

    [verbose] Log file: gitmap-debug-2025-03-10T14-30.log
    Cloning from .gitmap/output/gitmap.txt...
    [1/5] Cloning https://github.com/user/my-api.git... done
    ...
    Clone complete: 5 succeeded, 0 failed

## See Also

- [scan](scan.md) — Scan directories to generate output files
- [pull](pull.md) — Pull individual or grouped repos
- [desktop-sync](desktop-sync.md) — Sync repos to GitHub Desktop
- [clone-next](clone-next.md) — Clone next version of a repo
