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
gitmap reclone                                            # auto-pickup .gitmap/output/gitmap.json (then .csv)
gitmap reclone  <file>                                    # dry-run (default)
gitmap reclone  <file> --execute                          # actually clone
gitmap reclone  --manifest <path>                         # explicit manifest (JSON or CSV)
gitmap reclone  --manifest <path> --execute               # explicit + execute
gitmap reclone  --scan-root <dir> --execute               # auto-pickup from <dir>/.gitmap/output/
gitmap reclone  --execute                                 # auto-pickup + execute
gitmap reclone  <file> --mode ssh --execute               # use SSH URLs
gitmap rec      <file> --execute                          # short alias
gitmap clone-now <file> --execute                         # legacy alias (kept forever)
gitmap cnow     <file> --execute                          # legacy short alias
gitmap relclone <file> --execute                          # legacy alias
gitmap rc       <file> --execute                          # legacy short alias
```

## Source resolution

`reclone` picks the input file in this priority order:

1. `--manifest <path>`  — explicit, highest priority. JSON or CSV;
   format is auto-detected from the extension (override with `--format`).
2. Positional `<file>`  — legacy form, kept for back-compat.
3. Auto-pickup          — searches `<root>/.gitmap/output/gitmap.json`
   then `<root>/.gitmap/output/gitmap.csv`. `<root>` defaults to the
   current directory and can be redirected with `--scan-root <dir>`.

Passing **both** `--manifest` and a positional `<file>` is a usage error
(exit `2`) so the chosen artifact is unambiguous. `--scan-root` is only
consulted by the auto-pickup branch — it is silently ignored when an
explicit path is supplied.


## Auto-pickup

When `<file>` and `--manifest` are both omitted, `reclone` looks for
a scan artifact under:

1. `<scan-root>/.gitmap/output/gitmap.json`  (preferred — richest schema)
2. `<scan-root>/.gitmap/output/gitmap.csv`   (fallback)

`<scan-root>` is the current directory by default, or the value of
`--scan-root <dir>` when supplied. The first match is used and its
path is echoed to stderr so the run stays reproducible. If neither
file exists, `reclone` exits with code `2` and tells you to run
`gitmap scan` first (or pass `--manifest` / a positional path).
Auto-pickup never walks parent or sibling directories.

## Safety prompt (existing destinations)

Before any `git clone` runs under `--execute`, `reclone` checks
whether any planned `RelativePath` already exists under `--cwd`
(default: current directory). If at least one does:

- **Interactive TTY**: lists up to 10 existing destinations + total
  count and prompts `Proceed with git clone against these
  destinations? [y/N]:`. Only `y` proceeds; anything else aborts
  with exit `2` and no side effects.
- **Non-TTY** (CI, piped, redirected stdin): refuses with exit `2`
  and tells you to pass `--yes`. There is no silent fallthrough —
  you must opt in explicitly.
- **`--yes` passed**: skips the prompt entirely.

The per-row `--on-exists` policy (`skip` / `update` / `force`)
still controls what actually happens to each existing directory;
this gate is a single high-level confirmation that fires *before*
any side effect, so an accidental `--on-exists force` against a
populated tree is impossible without explicit confirmation.

## Arguments

| Argument | Required | Description |
|---|---|---|
| `<file>` | no | Path to a `.json`, `.csv`, or `.txt` file produced by `gitmap scan` (typically under `.gitmap/output/`). When omitted, auto-pickup is used (see above). |

## Flags

| Flag | Default | Description |
|---|---|---|
| `--manifest` | (none) | Explicit path to the scan artifact (`.json` or `.csv`). Equivalent to the positional `<file>` argument; cannot be combined with one. |
| `--scan-root` | current dir | Directory whose `.gitmap/output/` is probed during auto-pickup. Lets you `reclone` a tree scanned elsewhere without `cd`. Ignored when `--manifest` or a positional `<file>` is given. |
| `--execute` | off | Actually run `git clone`. Without this flag, only the dry-run plan is printed. |
| `--yes` | off | Skip the pre-flight confirmation when destination folders already exist. **Required for non-interactive / CI runs** — without a TTY and without `--yes`, `reclone --execute` exits `2` rather than block on stdin. The `--on-exists` policy still applies per row. |
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
- `2` — bad CLI usage (missing `<file>`, invalid flag value), OR the safety prompt was declined / refused (existing destinations + non-TTY without `--yes`).
