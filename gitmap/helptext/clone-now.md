# gitmap clone-now

Re-run `git clone` against the JSON / CSV / text artifacts produced by
`gitmap scan`, honoring the recorded folder structure and a user-
selected SSH/HTTPS mode.

## Synopsis

```
gitmap clone-now <file>                            # dry-run (default)
gitmap clone-now <file> --execute                  # actually clone
gitmap clone-now <file> --mode ssh --execute       # use SSH URLs
gitmap cnow <file> --execute                       # short alias
```

## Arguments

| Argument | Required | Description |
|---|---|---|
| `<file>` | yes | Path to a `.json`, `.csv`, or `.txt` file produced by `gitmap scan` (typically under `.gitmap/output/`). |

## Flags

| Flag | Default | Description |
|---|---|---|
| `--execute` | off | Actually run `git clone`. Without this flag, only the dry-run plan is printed. |
| `--quiet` | off | Suppress per-row progress lines. The end-of-batch summary still prints. |
| `--mode` | `https` | URL mode to clone with: `https` or `ssh`. Falls back to the other mode if the preferred URL is missing on a row. |
| `--format` | (auto) | Force input format: `json`, `csv`, or `text`. Default: detected from the file extension. |
| `--cwd` | (current dir) | Working directory each `git clone` runs in. Useful for re-creating a tree under a fresh root. |
| `--output <mode>` | (off) | Per-repo summary format. `terminal` = standardized branch/from/to/command block on **stdout**, streamed immediately before each row's `git clone`. Git progress and the batch summary stay on **stderr**. Empty (default) keeps the legacy output. |
| `--help` | off | Print this help and exit. |

## Output streams (`--output terminal`)

| Stream | Content |
|---|---|
| **stdout** | One `RepoTermBlock` per row (index, name, branch + source, original URL, target URL, exact `git clone` command). Streamed: each block prints right before that row's clone starts. Dry-run prints all blocks upfront (no clone to interleave with). |
| **stderr** | `git clone` progress, the `[i/N] status url -> dest` per-row line, the final summary, and any warnings. |

Redirect example: `gitmap clone-now repos.json --execute --output terminal > previews.txt 2> progress.log`.

## Input formats

`clone-now` consumes the same files `gitmap scan` writes:

- **JSON** -- `repos.json`: array of `ScanRecord` objects (`repoName`,
  `httpsUrl`, `sshUrl`, `branch`, `relativePath`, ...).
- **CSV** -- `repos.csv`: header row + one data row per repo (the
  format produced by `gitmap scan`'s default CSV output).
- **Text** -- `repos.txt` (or any `.txt` / unknown extension): one
  `git clone <url> [dest]` line per repo. Blank lines and `#`
  comments are ignored. Branch info is not preserved in this format
  -- use JSON or CSV if you need branch pinning.

## Behavior

- **Folder structure is preserved**: each row clones into its
  recorded `relativePath` (relative to `--cwd` or the current
  directory), so the destination tree mirrors the original layout.
- **Mode selection**: `--mode https` (default) clones via the
  recorded HTTPS URL; `--mode ssh` uses the SSH URL. If the chosen
  URL is empty on a row, `clone-now` falls back to the other one
  rather than skipping.
- **Idempotent**: a destination directory that already exists and is
  non-empty is reported as `skipped`, not re-cloned. Re-running the
  same input is safe.
- **Sequential**: rows are cloned in input order. Parallel fan-out
  is a future addition.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Dry-run completed successfully, or every row finished `ok` / `skipped` on `--execute`. |
| `1` | File open / parse error, or any row failed during `--execute`. |
| `2` | Bad CLI usage (missing `<file>`, invalid `--mode`, invalid `--format`). |

## Examples

Dry-run a previous scan:

```
gitmap clone-now .gitmap/output/repos.json
```

Re-clone everything under a mirror root using SSH:

```
gitmap clone-now .gitmap/output/repos.csv --mode ssh --cwd ./mirror --execute
```

Force the text-format parser on a renamed file:

```
gitmap clone-now my-repos.list --format text --execute
```
