# gitmap reclone

Re-run `git clone` against the JSON / CSV / text artifacts produced by
`gitmap scan`, honoring the recorded folder structure and a user-
selected SSH/HTTPS mode.

> **`reclone` vs `clone`** — different commands, different inputs.
>
> | Command | Input | Use when |
> |---|---|---|
> | `gitmap clone <url> [folder]` | A single repo URL | You want to clone (or re-clone) one repo from a URL. |
> | `gitmap reclone <file>` | A `gitmap scan` artifact (JSON / CSV / TXT) | You want to round-trip an entire previously-scanned tree at its recorded relative paths. |
>
> If you're generating clone *commands* without running them, that's
> still `gitmap scan` — `reclone` is the side that consumes those
> artifacts and actually re-creates the tree.

## Synopsis

```
gitmap reclone  <file>                            # dry-run (default)
gitmap reclone  <file> --execute                  # actually clone
gitmap reclone  <file> --mode ssh --execute       # use SSH URLs
gitmap rec      <file> --execute                  # short alias
gitmap clone-now <file> --execute                 # legacy alias (kept forever)
gitmap cnow     <file> --execute                  # legacy short alias
gitmap relclone <file> --execute                  # legacy alias
gitmap rc       <file> --execute                  # legacy short alias
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
| `--format` | auto | Force input format: `json`, `csv`, or `text`. Default auto-detects from the file extension. |
| `--cwd` | current dir | Working directory `git clone` runs in. Use to re-create the tree under a fresh root. |
| `--on-exists` | `skip` | Behavior when target already exists: `skip` (no-op when repo+branch match), `update` (fetch + checkout to align), `force` (remove target and re-clone — destructive). |
| `--max-concurrency` | auto | Worker count for parallel re-clones. `0` = `runtime.NumCPU()`, `1` = sequential. |

## Aliases

`reclone` is the canonical name. The following spellings dispatch to
the exact same command and flag set, kept for backward compatibility:

- `rec`
- `clone-now`, `cnow`
- `relclone`, `rc`

## Examples

```
# Round-trip a previously scanned tree under a fresh root.
gitmap reclone .gitmap/output/repos.json --cwd ./mirror --execute

# Re-align an existing checkout with the recorded URL/branch.
gitmap reclone .gitmap/output/repos.csv --on-exists update --execute

# Inspect what would happen, with no side effects.
gitmap reclone .gitmap/output/repos.json
```

## Exit codes

- `0` — dry-run completed, OR every row was ok/skipped on `--execute`.
- `1` — file open / parse error, OR any row failed on `--execute`.
- `2` — bad CLI usage (missing `<file>` or invalid flag value).
