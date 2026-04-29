# CI Hardening Session — 2026-04-29

> Session focused on tightening the GitHub Actions `lint` + compile pipeline. All changes scoped to `.github/workflows/ci.yml` (no Go source changes).

## ✅ Done

1. **Compile gate** — Added a `go test ./...` step after code changes; any typecheck/build error now fails CI before the matrix tests run.
2. **Build/test caching** — Enabled `GOMODCACHE` + `GOCACHE` reuse via `actions/setup-go@v5` cache + explicit `actions/cache` keys for `go.sum` hash. Speeds up the compile gate and downstream jobs.
3. **gofmt check** — Added a step that runs `gofmt -l` on all `.go` files; fails (with file list) on any unformatted file.
4. **goimports check** — Added a `goimports` step pinned to `golang.org/x/tools/cmd/goimports@v0.24.0`. Reads `LOCAL_PREFIX` dynamically from `go.mod`, runs `-l` for detection then `-d` for diff output, prints copy-pasteable fix command. Positioned between `gofmt` and `go vet`.
5. **golangci-lint strict gate** — Updated `golangci/golangci-lint-action@v6` step args to include `--issues-exit-code=1`. Pinned version remains `v1.64.8`, timeout `5m`. Working dir `gitmap`.
6. **Regression-guard semantics audit** — Documented that the "fail only on new issues" contract is split:
   - **Hard floor (zero-tolerance, ignores baseline):** `lint-regression-guard` job for `unused` + `gosec G115`, via `.github/scripts/check-lint-regressions.sh`.
   - **Baseline-diff (fail only on NEW):** `lint-baseline-diff` job (full JSON diff via `.github/scripts/lint-diff.py`) and per-linter sub-steps for `misspell`, `gocritic`, `exhaustive` via `.github/scripts/check-single-linter-diff.sh`.
   - Baseline cache: `golangci-baseline-main-…`, refreshed only on successful pushes to `main`.

## ⏳ Pending / Open

- 🚫 Decision needed from user: convert `unused` + `G115` from hard-floor to baseline-diff semantics, OR rename the job/docs to clearly say "hard floor" rather than "regression guard". Currently left as-is awaiting input.
- ⏳ Verify CI green on next push (no live run inspected this session).
- ⏳ Consider adding `goimports` and `gofmt` to the local `hooks/pre-commit` for parity with CI.

## Key Snippets

### goimports check
```yaml
- name: goimports check (import grouping + formatting)
  run: |
    go install golang.org/x/tools/cmd/goimports@v0.24.0
    GOIMPORTS="$(go env GOPATH)/bin/goimports"
    LOCAL_PREFIX="$(awk '/^module /{print $2; exit}' go.mod)"
    unformatted=$("$GOIMPORTS" -l -local "$LOCAL_PREFIX" .)
    if [ -n "$unformatted" ]; then
      echo "::error::The following .go files have goimports issues..."
      "$GOIMPORTS" -d -local "$LOCAL_PREFIX" $unformatted
      exit 1
    fi
```

### golangci-lint strict gate
```yaml
- name: golangci-lint (strict, fail on any error)
  if: needs.sha-check.outputs.already-built != 'true'
  uses: golangci/golangci-lint-action@v6
  with:
    version: v1.64.8
    working-directory: gitmap
    args: --timeout=5m --issues-exit-code=1
```

## Files Touched

- `.github/workflows/ci.yml` (all six steps above)

## Lessons / Anti-patterns

- **Pin every tool**: `goimports@v0.24.0`, `golangci-lint@v1.64.8`. Never `@latest` in CI.
- **Compute `-local` from `go.mod`**: avoids hardcoding the module path in CI.
- **Don't conflate floor vs diff**: `lint-regression-guard` is misleadingly named — half the linters it covers use a hard-floor (no baseline), half use baseline-diff. Either rename or unify before adding new linters to it.

## Next AI Pickup Point

If continuing CI work: ask the user the pending decision on `unused`/`G115` semantics, then either flip the script to baseline-diff or rename the job. Otherwise next logical step is wiring `gofmt`/`goimports` into `hooks/pre-commit`.
