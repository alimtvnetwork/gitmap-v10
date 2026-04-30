# 27 — `fix-repo` Command (Go-native)

## Purpose

`gitmap fix-repo` (alias `gitmap fr`) rewrites prior versioned-repo-name
tokens to the current version across all tracked text files. It is the
Go-native re-implementation of the existing `fix-repo.ps1` /
`fix-repo.sh` shell scripts, with byte-for-byte identical default
behavior, exit codes, and config schema. The PowerShell + Bash scripts
remain as bootstrap helpers; the Go command is the canonical entry
point and is shipped inside the `gitmap` binary.

## Synopsis

```
gitmap fix-repo [-2 | -3 | -5 | --all] [--dry-run] [--verbose] [--config <path>]
gitmap fr                                                       # short alias
```

PowerShell-style single-dash flags (`-2`, `-3`, `-5`, `-All`,
`-DryRun`, `-Verbose`, `-Config <p>`) are also accepted as aliases so
existing muscle memory continues to work.

## Behavior

1. Resolve repo identity from `git`:
   - `git rev-parse --show-toplevel` → repo root (else `E_NOT_A_REPO`).
   - `git config --get remote.origin.url` → remote URL (else `E_NO_REMOTE`).
   - Parse remote URL into `{host, owner, repo}` for HTTPS, SSH
     (`git@host:owner/repo`), and `ssh://` forms.
   - Split `repo` into `{base, version}` using regex `^(.+)-v(\d+)$`
     (else `E_NO_VERSION_SUFFIX`). Version must be ≥ 1
     (else `E_BAD_VERSION`).
2. Compute target span:
   - Default (no flag): last 2 prior versions.
   - `-2 | -3 | -5`: last N prior versions.
   - `--all`: every prior version `1..current-1`.
   - Targets clamp to `[max(1, current-span) .. current-1]`.
3. Load config (default `<repoRoot>/fix-repo.config.json`,
   override with `--config <path>`):
   - `ignoreDirs`: array of repo-relative directory prefixes to skip.
   - `ignorePatterns`: array of glob patterns. `**` matches across
     segments; `*` matches within a single segment; `?` matches one
     non-`/` char. Missing config file is non-fatal; explicit
     `--config <missing>` is fatal (`E_BAD_CONFIG`).
4. Enumerate tracked files via `git ls-files`. For each path:
   - Skip if matched by `ignoreDirs` / `ignorePatterns`.
   - Skip reparse points and files larger than 5 MiB.
   - Skip files whose extension is in the binary allow-list:
     `.png .jpg .jpeg .gif .webp .ico .pdf .zip .tar .gz .tgz .bz2
     .xz .7z .rar .woff .woff2 .ttf .otf .eot .mp3 .mp4 .mov .wav
     .ogg .webm .class .jar .so .dylib .dll .exe .pyc`.
   - Skip files whose first 8 KiB contain a NUL byte.
5. For each surviving file, replace every literal occurrence of
   `{base}-v{N}` (for each `N` in targets) with `{base}-v{current}`,
   guarded by a negative-lookahead so `{base}-v1` does **not** match
   inside `{base}-v10`. Counts every replacement.
6. In `--dry-run`, files are not written; counts are still reported.
7. Print header + summary identical to the PowerShell script:

   ```
   fix-repo  base=<base>  current=v<N>  mode=<mode>
   targets:  v1, v2, ...
   host:     <host>  owner=<owner>

   scanned: <S> files
   changed: <C> files (<R> replacements)
   mode:    write|dry-run
   ```

## Exit codes

| Code | Constant              | Meaning                                  |
|------|-----------------------|------------------------------------------|
| 0    | `ExitOk`              | Success.                                 |
| 2    | `ExitNotARepo`        | Not inside a git work tree.              |
| 3    | `ExitNoRemote`        | Missing or unparseable `origin` URL.     |
| 4    | `ExitNoVersionSuffix` | Repo name has no `-vN` suffix.           |
| 5    | `ExitBadVersion`      | Version ≤ 0.                             |
| 6    | `ExitBadFlag`         | Unknown / conflicting CLI flag.          |
| 7    | `ExitWriteFailed`     | At least one file write failed.          |
| 8    | `ExitBadConfig`       | Explicit `--config` missing or invalid.  |

These match `fix-repo.ps1` and `fix-repo.sh` 1:1 so CI scripts that
inspect the exit code keep working when they switch invocation from
the script to the binary.

## Idempotency & safety

- Replacing `{base}-v{N}` with the same `{base}-v{current}` is
  idempotent: a clean repo with no prior versions yields
  `changed: 0`.
- The negative-lookahead (`(?!\d)`) prevents partial-token rewrites
  inside larger version numbers (`v10`, `v123`).
- Tracked-file enumeration via `git ls-files` automatically respects
  `.gitignore`, so untracked build output is never touched.
- The 5 MiB / NUL-byte / extension guards prevent corrupting
  binary assets that happen to be tracked.

## Naming + ownership

- Top-level command name: `fix-repo` (kebab-case, like `desktop-sync`).
- Short alias: `fr`.
- Constants live in `gitmap/constants/constants_fixrepo.go` (package-
  domain ownership, per the constants-ownership rule).
- Implementation is split across `gitmap/cmd/fixrepo*.go` files,
  each ≤ 200 lines, functions ≤ 15 lines, positive conditionals only,
  no swallowed errors (logged to `os.Stderr` via the standard format).

## Cross-references

- PowerShell script: `fix-repo.ps1` + `scripts/fix-repo/*.ps1`.
- POSIX script: `fix-repo.sh` + `scripts/fix-repo/*.sh`.
- Constants ownership rule: `spec/12-consolidated-guidelines/02-go-code-style.md`.
- Strictly-prohibited registry: `spec/03-general/10-strictly-prohibited.md`
  (no time/date in `readme.txt`, no manual edits to `.gitmap/release/`).
