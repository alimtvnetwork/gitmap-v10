# CI/CD Issue 02 — `lint-regression-guard` semantics are split (hard-floor vs baseline-diff)

## Pipeline
- **Workflow:** `.github/workflows/ci.yml`
- **Jobs:** `lint-regression-guard`, `lint-baseline-diff`
- **Scripts:** `.github/scripts/check-lint-regressions.sh`, `.github/scripts/lint-diff.py`, `.github/scripts/check-single-linter-diff.sh`

## Symptom
User asked to "verify the lint regression guard ignores the baseline and fails CI only on newly introduced golangci-lint issues." Verification revealed the contract is **not** uniform — the job mixes two enforcement models.

## Root Cause
Two distinct enforcement models live under the "regression guard" umbrella:

| Linter / Check | Model | Implementation |
|---|---|---|
| `unused` | **Hard floor** (no baseline) | `check-lint-regressions.sh` |
| `gosec G115` (integer-overflow) | **Hard floor** (no baseline) | `check-lint-regressions.sh` |
| `misspell` | **Baseline diff** (new only) | `check-single-linter-diff.sh` |
| `gocritic` | **Baseline diff** (new only) | `check-single-linter-diff.sh` |
| `exhaustive` | **Baseline diff** (new only) | `check-single-linter-diff.sh` |
| Full report | **Baseline diff** | `lint-diff.py` (in `lint-baseline-diff` job) |

Baseline is cached as `golangci-baseline-main-…`, refreshed only on successful pushes to `main`.

## Status
✅ Resolved (2026-04-29) — User chose Option (b). Job renamed `lint-regression-guard` → `lint-hard-floor` in `.github/workflows/ci.yml`. Comment block rewritten to clearly document that the primary step is hard-floor (`unused` + `G115`) and that the misspell/gocritic/exhaustive sub-steps remain baseline-diff (kept co-located for cache-key locality, with per-step comments stating their model). New job label: `Lint Hard Floor (unused + G115) + baseline-diff sub-steps`.

## Resolution Path (historical)
Two options presented to user:
- **(a)** Convert `unused` + `G115` to baseline-diff semantics (any new issue fails; existing baseline tolerated).
- **(b)** Rename `lint-regression-guard` → `lint-hard-floor` (or split into two jobs) so the job name reflects zero-tolerance enforcement and stops implying baseline-diff. ← **Selected.**

## Prevention
- When adding a new linter to CI, decide upfront: hard-floor or baseline-diff. Document the choice in the script header comment.
- Do NOT add new linters to `check-lint-regressions.sh` if you want baseline-diff semantics — use `check-single-linter-diff.sh` instead.

## Related
- Session memory: `.lovable/memory/workflow/04-ci-hardening-session.md`
